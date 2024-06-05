package usage

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . credentialsGetter
//counterfeiter:generate . Reporter

const apiBasePath = "/api/platform/v1/k8s-usage"

// ClusterDetails are the k8s usage details for the cluster.
type ClusterDetails struct {
	// Metadata contains the cluster metadata.
	Metadata Metadata `json:"metadata"`
	// PodDetails contain the details about the NGF Pod.
	PodDetails PodDetails `json:"pod_details"`
	// NodeCount is the count of Nodes in the cluster.
	NodeCount int64 `json:"node_count"`
}

// Metadata contains the cluster metadata.
type Metadata struct {
	// DisplayName is a user friendly resource name. It can be used to define
	// a longer, and less constrained, name for a resource.
	DisplayName string `json:"displayName"`
	// UID is the unique identifier for the cluster.
	UID string `json:"uid"`
}

// PodDetails contain the details about the NGF Pod.
type PodDetails struct {
	// CurrentPodsCount is the total count of NGF NGINX Plus Pods in the cluster.
	CurrentPodCounts CurrentPodsCount `json:"current_pod_counts"`
}

// CurrentPodsCount is the total count of NGF NGINX Plus Pods in the cluster.
type CurrentPodsCount struct {
	// PodCount is the current count of NGF NGINX Plus Pods in the cluster.
	PodCount int64 `json:"pod_count"`
	// DosCount is the count of Pods with NAP DOS enabled in the cluster. Not applicable for NGF,
	// but required as part of the payload.
	DosCount int64 `json:"dos_count"`
	// WafCount is the count of Pods with NAP WAF enabled in the cluster. Not applicable for NGF,
	// but required as part of the payload.
	WafCount int64 `json:"waf_count"`
}

// credentialsGetter get the credentials for NGINX Plus usage reporting.
type credentialsGetter interface {
	// GetCredentials returns the base64 encoded username and password from the Secret.
	GetCredentials() ([]byte, []byte)
}

// Reporter reports the NGINX Plus usage info to the provided collector.
type Reporter interface {
	Report(context.Context, ClusterDetails) error
}

// NIMReporter reports the NGINX Plus usage info to NGINX Instance Manager.
type NIMReporter struct {
	// credentials contains the credentials for the usage collector.
	credentials credentialsGetter
	// baseURL is the base server URL of the usage collector.
	baseURL *url.URL
	// insecureSkipVerify controls whether the client verifies the server cert. Used in testing.
	insecureSkipVerify bool
}

// NewNIMReporter creates a new NIM usage reporter.
func NewNIMReporter(
	credentials credentialsGetter,
	baseURL string,
	insecureSkipVerify bool,
) (*NIMReporter, error) {
	serverURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing usage server URL: %w", err)
	}

	return &NIMReporter{
		credentials:        credentials,
		baseURL:            serverURL,
		insecureSkipVerify: insecureSkipVerify,
	}, nil
}

// Report sends a PUT request with the provided data to the API endpoint configured in the Reporter.
// The clusterUID is used as the name in the API path.
func (r *NIMReporter) Report(ctx context.Context, data ClusterDetails) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling usage data: %w", err)
	}

	queryURL := r.baseURL.JoinPath(apiBasePath, data.Metadata.UID)
	bodyReader := bytes.NewReader(buf)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, queryURL.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("error creating usage API HTTP request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	username, password := r.credentials.GetCredentials()
	if username == nil || password == nil {
		return errors.New("username or password not set for NGINX Plus usage reporting; unable to send reports." +
			" Ensure that the usage Secret exists and the username and password are set")
	}
	req.SetBasicAuth(string(username), string(password))

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: r.insecureSkipVerify, //nolint:gosec // used for testing
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending usage report request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read the response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response: %v; response body: %v", resp.StatusCode, string(body))
	}

	return nil
}
