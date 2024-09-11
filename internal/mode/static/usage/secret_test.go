package usage

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSet(t *testing.T) {
	t.Parallel()
	store := NewUsageSecret()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "custom",
		},
	}

	g := NewWithT(t)
	g.Expect(store.secret).To(BeNil())

	store.Set(secret)
	g.Expect(store.secret).To(Equal(secret))
}

func TestDelete(t *testing.T) {
	t.Parallel()
	store := NewUsageSecret()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "custom",
		},
	}

	g := NewWithT(t)
	store.Set(secret)
	g.Expect(store.secret).To(Equal(secret))

	store.Delete()
	g.Expect(store.secret).To(BeNil())
}

func TestGetCredentials(t *testing.T) {
	t.Parallel()
	store := NewUsageSecret()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret",
			Namespace: "custom",
		},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}

	g := NewWithT(t)

	user, pass := store.GetCredentials()
	g.Expect(user).To(BeNil())
	g.Expect(pass).To(BeNil())

	store.Set(secret)

	user, pass = store.GetCredentials()
	g.Expect(user).To(Equal([]byte("user")))
	g.Expect(pass).To(Equal([]byte("pass")))
}
