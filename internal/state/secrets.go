package state

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecretStore
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SecretMemoryManager

// tlsSecretFileMode defines the default file mode for files with TLS Secrets.
const tlsSecretFileMode = 0o600

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

func NewSecretStore() *SecretStoreImpl {
	return &SecretStoreImpl{
		secrets: make(map[types.NamespacedName]*Secret),
	}
}

func (s SecretStoreImpl) Upsert(secret *apiv1.Secret) {
	nsname := types.NamespacedName{
		Namespace: secret.Namespace,
		Name:      secret.Name,
	}

	valid := isSecretValid(secret)
	s.secrets[nsname] = &Secret{Secret: secret, Valid: valid}
}

func (s SecretStoreImpl) Delete(nsname types.NamespacedName) {
	delete(s.secrets, nsname)
}

func (s SecretStoreImpl) Get(nsname types.NamespacedName) *Secret {
	return s.secrets[nsname]
}

type SecretMemoryManager interface {
	// Store stores the secret in memory so that it can be written to disk before reloading NGINX.
	// Returns the path to the secret and an error if the secret does not exist in the cache or the secret is invalid.
	Store(nsname types.NamespacedName) (string, error)
	// WriteAllStoredSecrets writes all stored secrets to disk.
	WriteAllStoredSecrets() error
}

type SecretDiskMemoryManager struct {
	storedSecrets   map[types.NamespacedName]storedSecret
	secretCache     SecretStore
	secretDirectory string
}

type storedSecret struct {
	secret *apiv1.Secret
	path   string
}

func NewSecretDiskMemoryManager(secretDirectory string, secretStore SecretStore) *SecretDiskMemoryManager {
	return &SecretDiskMemoryManager{
		storedSecrets:   make(map[types.NamespacedName]storedSecret),
		secretCache:     secretStore,
		secretDirectory: secretDirectory,
	}
}

func (s *SecretDiskMemoryManager) Store(nsname types.NamespacedName) (string, error) {
	secret := s.secretCache.Get(nsname)
	if secret == nil {
		return "", fmt.Errorf("secret %s does not exist", nsname)
	}

	if !secret.Valid {
		return "", fmt.Errorf("secret %s is not valid; must be of type %s and contain a valid X509 key pair", nsname, apiv1.SecretTypeTLS)
	}

	ss := storedSecret{
		secret: secret.Secret,
		path:   path.Join(s.secretDirectory, generateFilepathForSecret(nsname)),
	}

	s.storedSecrets[nsname] = ss

	return ss.path, nil
}

func (s *SecretDiskMemoryManager) WriteAllStoredSecrets() error {
	// Remove all existing secrets from secrets directory
	dir, err := ioutil.ReadDir(s.secretDirectory)
	if err != nil {
		return fmt.Errorf("failed to remove all secrets from %s: %w", s.secretDirectory, err)
	}

	for _, d := range dir {
		filepath := path.Join(s.secretDirectory, d.Name())
		if err := os.Remove(filepath); err != nil {
			return fmt.Errorf("failed to remove secret %s: %w", filepath, err)
		}
	}

	// Write all secrets to secrets directory
	for nsname, ss := range s.storedSecrets {

		file, err := os.Create(ss.path)
		if err != nil {
			return fmt.Errorf("failed to create file %s for secret %s: %w", ss.path, nsname, err)
		}

		if err = file.Chmod(tlsSecretFileMode); err != nil {
			return fmt.Errorf("failed to change mode of file %s for secret %s: %w", ss.path, nsname, err)
		}

		contents := generateCertAndKeyFileContent(ss.secret)

		_, err = file.Write(contents)
		if err != nil {
			return fmt.Errorf("failed to write secret %s to file %s: %w", nsname, ss.path, err)
		}

	}

	// reset stored secrets
	s.storedSecrets = make(map[types.NamespacedName]storedSecret)

	return nil
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
