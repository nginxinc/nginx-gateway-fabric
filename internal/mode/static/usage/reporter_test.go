package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
)

func TestReport(t *testing.T) {
	g := NewWithT(t)

	data := ClusterDetails{
		Metadata: Metadata{
			UID:         "12345abcde",
			DisplayName: "my-cluster",
		},
		NodeCount: 9,
		PodDetails: PodDetails{
			CurrentPodCounts: CurrentPodsCount{
				PodCount: 12,
			},
		},
	}

	secret := &v1.Secret{
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}

	store := NewUsageSecret()
	store.Set(secret)

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqData ClusterDetails
			g.Expect(json.NewDecoder(r.Body).Decode(&reqData)).To(Succeed())
			g.Expect(reqData).To(Equal(data))

			g.Expect(r.URL.Path).To(Equal(fmt.Sprintf("%s/%s", apiBasePath, data.Metadata.UID)))
			g.Expect(r.Method).To(Equal(http.MethodPut))
			user, pass, ok := r.BasicAuth()
			g.Expect(ok).To(BeTrue())
			g.Expect(user).To(Equal("user"))
			g.Expect(pass).To(Equal("pass"))

			contentType, ok := r.Header["Content-Type"]
			g.Expect(ok).To(BeTrue())
			g.Expect(contentType[0]).To(Equal("application/json"))

			w.WriteHeader(http.StatusOK)
		}),
	)
	defer server.Close()

	insecureSkipVerify := false
	reporter, err := NewNIMReporter(store, server.URL, insecureSkipVerify)
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(reporter.Report(context.Background(), data)).To(Succeed())
}

func TestReport_NoCredentials(t *testing.T) {
	g := NewWithT(t)
	insecureSkipVerify := false
	reporter, err := NewNIMReporter(NewUsageSecret(), "", insecureSkipVerify)
	g.Expect(err).ToNot(HaveOccurred())

	err = reporter.Report(context.Background(), ClusterDetails{})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("username or password not set"))
}

func TestReport_ServerError(t *testing.T) {
	g := NewWithT(t)

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	)
	defer server.Close()

	secret := &v1.Secret{
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}

	store := NewUsageSecret()
	store.Set(secret)

	insecureSkipVerify := false
	reporter, err := NewNIMReporter(store, server.URL, insecureSkipVerify)
	g.Expect(err).ToNot(HaveOccurred())

	err = reporter.Report(context.Background(), ClusterDetails{})
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("non-200 response"))
}
