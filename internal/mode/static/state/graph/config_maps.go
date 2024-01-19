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

// CaCertConfigMap represents a ConfigMap resource that holds CA Cert data.
type CaCertConfigMap struct {
	// Source holds the actual ConfigMap resource. Can be nil if the ConfigMap does not exist.
	Source *apiv1.ConfigMap
	// CACert holds the actual CA Cert data.
	CACert []byte
}

type caCertConfigMapEntry struct {
	// err holds the corresponding error if the ConfigMap is invalid or does not exist.
	err             error
	caCertConfigMap CaCertConfigMap
}

// configMapResolver wraps the cluster ConfigMaps so that they can be resolved (includes validation). All resolved
// ConfigMaps are saved to be used later.
type configMapResolver struct {
	clusterConfigMaps  map[types.NamespacedName]*apiv1.ConfigMap
	resolvedConfigMaps map[types.NamespacedName]*caCertConfigMapEntry
}

func newConfigMapResolver(configMaps map[types.NamespacedName]*apiv1.ConfigMap) *configMapResolver {
	return &configMapResolver{
		clusterConfigMaps:  configMaps,
		resolvedConfigMaps: make(map[types.NamespacedName]*caCertConfigMapEntry),
	}
}

func (r *configMapResolver) resolve(nsname types.NamespacedName) error {
	if s, resolved := r.resolvedConfigMaps[nsname]; resolved {
		return s.err
	}

	cm, exist := r.clusterConfigMaps[nsname]

	var validationErr error
	var caCert []byte

	if !exist {
		validationErr = errors.New("configMap does not exist")
	}

	if exist {
		if cm.Data != nil {
			if _, exists := cm.Data[CAKey]; exists {
				validationErr = validateCA([]byte(cm.Data[CAKey]))
				caCert = []byte(cm.Data[CAKey])
			}
		}
		if cm.BinaryData != nil {
			if _, exists := cm.BinaryData[CAKey]; exists {
				validationErr = validateCA(cm.BinaryData[CAKey])
				caCert = cm.BinaryData[CAKey]
			}
		}
		if len(caCert) == 0 {
			validationErr = fmt.Errorf("configMap does not have the data or binaryData field %v", CAKey)
		}
	}

	r.resolvedConfigMaps[nsname] = &caCertConfigMapEntry{
		caCertConfigMap: CaCertConfigMap{
			Source: cm,
			CACert: caCert,
		},
		err: validationErr,
	}

	return validationErr
}

func (r *configMapResolver) getResolvedConfigMaps() map[types.NamespacedName]*CaCertConfigMap {
	if len(r.resolvedConfigMaps) == 0 {
		return nil
	}

	resolved := make(map[types.NamespacedName]*CaCertConfigMap)

	for nsname, entry := range r.resolvedConfigMaps {
		// create iteration variable inside the loop to fix implicit memory aliasing
		caCertConfigMap := entry.caCertConfigMap
		resolved[nsname] = &caCertConfigMap
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
