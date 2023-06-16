package graph

import (
	"errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Secret struct {
	Source *apiv1.Secret
	Valid  bool
}

type secretResolver struct {
	secrets map[types.NamespacedName]*apiv1.Secret
	loaded  map[types.NamespacedName]*Secret
}

func newSecretResolver(secrets map[types.NamespacedName]*apiv1.Secret) *secretResolver {
	return &secretResolver{
		secrets: secrets,
		loaded:  make(map[types.NamespacedName]*Secret),
	}
}

func (r *secretResolver) Resolve(nsname types.NamespacedName) error {
	if _, ok := r.loaded[nsname]; ok {
		return nil
	}

	secret, ok := r.secrets[nsname]
	if !ok {
		r.loaded[nsname] = &Secret{
			Valid: false,
		}
		return errors.New("secret not found")
	}

	// TODO: validate secret

	r.loaded[nsname] = &Secret{
		Source: secret,
		Valid:  true,
	}

	return nil
}

func (r *secretResolver) ResolvedSecrets() map[types.NamespacedName]*Secret {
	return r.loaded
}
