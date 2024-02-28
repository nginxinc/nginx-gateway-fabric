package usage

import (
	"sync"

	v1 "k8s.io/api/core/v1"
)

// Secret implements the SecretStorer interface.
type Secret struct {
	secret *v1.Secret
	lock   *sync.Mutex
}

// NewUsageSecret creates a new Secret wrapper.
func NewUsageSecret() *Secret {
	return &Secret{
		lock: &sync.Mutex{},
	}
}

// Set stores the updated Secre.
func (s *Secret) Set(secret *v1.Secret) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.secret = secret
}

// Delete nullifies the Secret value.
func (s *Secret) Delete() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.secret = nil
}

// GetCredentials returns the base64 encoded username and password from the Secret.
func (s *Secret) GetCredentials() ([]byte, []byte) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.secret != nil {
		return s.secret.Data["username"], s.secret.Data["password"]
	}

	return nil, nil
}
