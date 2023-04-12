package secrets

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"path"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecretStore
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RequestManager

// SecretStore stores secrets.
type SecretStore interface {
	// Upsert upserts the secret into the store.
	Upsert(secret *apiv1.Secret)
	// Delete deletes the secret from the store.
	Delete(nsname types.NamespacedName)
	// Get gets the secret from the store.
	Get(nsname types.NamespacedName) *Secret
}

type SecretStoreImpl struct {
	secrets map[types.NamespacedName]*Secret
}

// Secret is the internal representation of a Kubernetes Secret.
type Secret struct {
	// Secret is the Kubernetes Secret object.
	Secret *apiv1.Secret
	// Valid is whether the Kubernetes Secret is valid.
	Valid bool
}

// File represents a secret as a file. Contains the file name and the file contents.
type File struct {
	Name     string
	Contents []byte
}

func NewSecretStore() *SecretStoreImpl {
	return &SecretStoreImpl{
		secrets: make(map[types.NamespacedName]*Secret),
	}
}

func (s *SecretStoreImpl) Upsert(secret *apiv1.Secret) {
	nsname := types.NamespacedName{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	valid := isSecretValid(secret)
	s.secrets[nsname] = &Secret{Secret: secret, Valid: valid}
}

func (s *SecretStoreImpl) Delete(nsname types.NamespacedName) {
	delete(s.secrets, nsname)
}

func (s *SecretStoreImpl) Get(nsname types.NamespacedName) *Secret {
	return s.secrets[nsname]
}

// RequestManager manages secrets that are requested by Gateway resources.
type RequestManager interface {
	// Request marks the secret as requested so that it can be written to disk before reloading NGINX.
	// Returns the path to the secret if it exists.
	// Returns an error if the secret does not exist in the secret store or the secret is invalid.
	Request(nsname types.NamespacedName) (string, error)
	// GetAndResetRequestedSecrets returns all request secrets as Files and resets the requested secrets.
	GetAndResetRequestedSecrets() []File
}

// RequestManagerImpl is the implementation of RequestManager.
// FIXME(kate-osborn): Is it necessary to make this concurrent-safe?
type RequestManagerImpl struct {
	requestedSecrets map[types.NamespacedName]requestedSecret
	secretStore      SecretStore
	secretDirectory  string
}

type requestedSecret struct {
	secret *apiv1.Secret
	path   string
}

func NewRequestManagerImpl(
	secretDirectory string,
	secretStore SecretStore,
) *RequestManagerImpl {
	return &RequestManagerImpl{
		requestedSecrets: make(map[types.NamespacedName]requestedSecret),
		secretStore:      secretStore,
		secretDirectory:  secretDirectory,
	}
}

func (s *RequestManagerImpl) Request(nsname types.NamespacedName) (string, error) {
	secret := s.secretStore.Get(nsname)
	if secret == nil {
		return "", fmt.Errorf("secret %s does not exist", nsname)
	}

	if !secret.Valid {
		return "", fmt.Errorf(
			"secret %s is not valid; must be of type %s and contain a valid X509 key pair",
			nsname,
			apiv1.SecretTypeTLS,
		)
	}

	ss := requestedSecret{
		secret: secret.Secret,
		path:   path.Join(s.secretDirectory, generateFilepathForSecret(nsname)),
	}

	s.requestedSecrets[nsname] = ss

	return ss.path, nil
}

func (s *RequestManagerImpl) GetAndResetRequestedSecrets() []File {
	files := make([]File, 0, len(s.requestedSecrets))
	for _, secret := range s.requestedSecrets {
		files = append(files, File{
			Name:     secret.path,
			Contents: generateCertAndKeyFileContent(secret.secret),
		})
	}

	s.requestedSecrets = make(map[types.NamespacedName]requestedSecret)

	return files
}

func isSecretValid(secret *apiv1.Secret) bool {
	if secret.Type != apiv1.SecretTypeTLS {
		return false
	}

	// A TLS Secret is guaranteed to have these data fields.
	_, err := tls.X509KeyPair(secret.Data[apiv1.TLSCertKey], secret.Data[apiv1.TLSPrivateKeyKey])

	return err == nil
}

func generateCertAndKeyFileContent(secret *apiv1.Secret) []byte {
	var res bytes.Buffer

	res.Write(secret.Data[apiv1.TLSCertKey])
	res.WriteString("\n")
	res.Write(secret.Data[apiv1.TLSPrivateKeyKey])

	return res.Bytes()
}

func generateFilepathForSecret(nsname types.NamespacedName) string {
	return nsname.Namespace + "_" + nsname.Name
}
