package graph

import (
	"crypto/tls"
	"errors"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Secret represents a Secret resource.
type Secret struct {
	// Source holds the actual Secret resource. Can be nil if the Secret does not exist.
	Source *apiv1.Secret
}

type secretEntry struct {
	Secret
	// err holds the corresponding error if the Secret is invalid or does not exist.
	err error
}

// secretResolver wraps the cluster Secrets so that they can be resolved (includes validation). All resolved
// Secrets are saved to be used later.
type secretResolver struct {
	clusterSecrets  map[types.NamespacedName]*apiv1.Secret
	resolvedSecrets map[types.NamespacedName]*secretEntry
}

func newSecretResolver(secrets map[types.NamespacedName]*apiv1.Secret) *secretResolver {
	return &secretResolver{
		clusterSecrets:  secrets,
		resolvedSecrets: make(map[types.NamespacedName]*secretEntry),
	}
}

func (r *secretResolver) resolve(nsname types.NamespacedName) error {
	if s, resolved := r.resolvedSecrets[nsname]; resolved {
		return s.err
	}

	secret, exist := r.clusterSecrets[nsname]

	var validationErr error

	if !exist {
		validationErr = errors.New("secret does not exist")
	} else if secret.Type != apiv1.SecretTypeTLS {
		validationErr = fmt.Errorf("secret type must be %q not %q", apiv1.SecretTypeTLS, secret.Type)
	} else {
		// A TLS Secret is guaranteed to have these data fields.
		_, err := tls.X509KeyPair(secret.Data[apiv1.TLSCertKey], secret.Data[apiv1.TLSPrivateKeyKey])
		if err != nil {
			validationErr = fmt.Errorf("TLS secret is invalid: %w", err)
		}
	}

	r.resolvedSecrets[nsname] = &secretEntry{
		Secret: Secret{
			Source: secret,
		},
		err: validationErr,
	}

	return validationErr
}

func (r *secretResolver) getResolvedSecrets() map[types.NamespacedName]*Secret {
	if len(r.resolvedSecrets) == 0 {
		return nil
	}

	resolved := make(map[types.NamespacedName]*Secret)

	for nsname, entry := range r.resolvedSecrets {
		resolved[nsname] = &entry.Secret
	}

	return resolved
}
