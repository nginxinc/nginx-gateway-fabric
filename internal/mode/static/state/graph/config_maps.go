package graph

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const CAKey = "ca.crt"

// ConfigMap represents a ConfigMap resource.
type ConfigMap struct {
	// Source holds the actual ConfigMap resource. Can be nil if the ConfigMap does not exist.
	Source *apiv1.ConfigMap
}

type configMapEntry struct {
	ConfigMap
	// err holds the corresponding error if the ConfigMap is invalid or does not exist.
	err error
}

// configMapResolver wraps the cluster ConfigMaps so that they can be resolved (includes validation). All resolved
// ConfigMaps are saved to be used later.
type configMapResolver struct {
	clusterConfigMaps  map[types.NamespacedName]*apiv1.ConfigMap
	resolvedConfigMaps map[types.NamespacedName]*configMapEntry
}

func newConfigMapResolver(configMaps map[types.NamespacedName]*apiv1.ConfigMap) *configMapResolver {
	return &configMapResolver{
		clusterConfigMaps:  configMaps,
		resolvedConfigMaps: make(map[types.NamespacedName]*configMapEntry),
	}
}

func (r *configMapResolver) resolve(nsname types.NamespacedName) error {
	if s, resolved := r.resolvedConfigMaps[nsname]; resolved {
		return s.err
	}

	cm, exist := r.clusterConfigMaps[nsname]

	var validationErr error

	if !exist {
		validationErr = errors.New("configMap does not exist")
	}

	if exist {

		var caCrtPresent bool

		if cm.Data != nil {
			if _, exists := cm.Data[CAKey]; exists {
				validationErr = validateCA([]byte(cm.Data[CAKey]))
				caCrtPresent = true
			}
		}

		if cm.BinaryData != nil {
			if _, exists := cm.BinaryData[CAKey]; exists {
				validationErr = validateCA(cm.BinaryData[CAKey])
				caCrtPresent = true
			}
		}

		if !caCrtPresent {
			validationErr = fmt.Errorf("configMap does not have the data or binaryData field %v", CAKey)
		}
	}

	r.resolvedConfigMaps[nsname] = &configMapEntry{
		ConfigMap: ConfigMap{
			Source: cm,
		},
		err: validationErr,
	}

	return validationErr
}

func (r *configMapResolver) getResolvedConfigMaps() map[types.NamespacedName]*ConfigMap {
	if len(r.resolvedConfigMaps) == 0 {
		return nil
	}

	resolved := make(map[types.NamespacedName]*ConfigMap)

	for nsname, entry := range r.resolvedConfigMaps {
		// create iteration variable inside the loop to fix implicit memory aliasing
		configMap := entry.ConfigMap
		resolved[nsname] = &configMap
	}

	return resolved
}

// validateCA validates the ca.crt entry in the ConfigMap. If it is valid, the function returns nil.
func validateCA(caData []byte) error {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(caData)))
	_, err := base64.StdEncoding.Decode(data, caData)
	if err != nil {
		data = caData
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("the data field %s must hold a valid CERTIFICATE PEM block", CAKey)
	}
	if block.Type != "CERTIFICATE" {
		return fmt.Errorf("the data field %s must hold a valid CERTIFICATE PEM block, but got '%s'", CAKey, block.Type)
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to validate certificate: %w", err)
	}

	return nil
}
