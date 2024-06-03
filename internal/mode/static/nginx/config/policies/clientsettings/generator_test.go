package clientsettings_test

import (
	"testing"

	. "github.com/onsi/gomega"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies/clientsettings"
)

func TestGenerate(t *testing.T) {
	maxSize := helpers.GetPointer[ngfAPI.Size]("10m")
	bodyTimeout := helpers.GetPointer[ngfAPI.Duration]("600ms")
	keepaliveRequests := helpers.GetPointer[int32](900)
	keepaliveTime := helpers.GetPointer[ngfAPI.Duration]("50s")
	keepaliveServerTimeout := helpers.GetPointer[ngfAPI.Duration]("30s")
	keepaliveHeaderTimeout := helpers.GetPointer[ngfAPI.Duration]("60s")

	tests := []struct {
		name       string
		policy     policies.Policy
		expStrings []string
	}{
		{
			name: "body max size populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						MaxSize: maxSize,
					},
				},
			},
			expStrings: []string{
				"client_max_body_size 10m;",
			},
		},
		{
			name: "body timeout populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						Timeout: bodyTimeout,
					},
				},
			},
			expStrings: []string{
				"client_body_timeout 600ms",
			},
		},
		{
			name: "keepalive requests populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Requests: keepaliveRequests,
					},
				},
			},
			expStrings: []string{
				"keepalive_requests 900;",
			},
		},
		{
			name: "keepalive time populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Time: keepaliveTime,
					},
				},
			},
			expStrings: []string{
				"keepalive_time 50s;",
			},
		},
		{
			name: "keepalive timeout server populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Server: keepaliveServerTimeout,
						},
					},
				},
			},
			expStrings: []string{
				"keepalive_timeout 30s;",
			},
		},
		{
			name: "keepalive timeout server and header populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Server: keepaliveServerTimeout,
							Header: keepaliveHeaderTimeout,
						},
					},
				},
			},
			expStrings: []string{
				"keepalive_timeout 30s 60s;",
			},
		},
		{
			name: "keepalive timeout header populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Header: keepaliveHeaderTimeout,
						},
					},
				},
			},
			expStrings: []string{}, // header timeout is ignored if server timeout is not populated
		},
		{
			name: "all fields populated",
			policy: &ngfAPI.ClientSettingsPolicy{
				Spec: ngfAPI.ClientSettingsPolicySpec{
					Body: &ngfAPI.ClientBody{
						MaxSize: maxSize,
						Timeout: bodyTimeout,
					},
					KeepAlive: &ngfAPI.ClientKeepAlive{
						Requests: keepaliveRequests,
						Time:     keepaliveTime,
						Timeout: &ngfAPI.ClientKeepAliveTimeout{
							Server: keepaliveServerTimeout,
							Header: keepaliveHeaderTimeout,
						},
					},
				},
			},
			expStrings: []string{
				"client_max_body_size 10m;",
				"client_body_timeout 600ms",
				"keepalive_requests 900;",
				"keepalive_time 50s;",
				"keepalive_timeout 30s 60s;",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			generator := clientsettings.NewGenerator()

			resFiles := generator.GenerateForServer([]policies.Policy{test.policy}, http.Server{})
			g.Expect(resFiles).To(HaveLen(1))

			for _, str := range test.expStrings {
				g.Expect(string(resFiles[0].Content)).To(ContainSubstring(str))
			}

			resFiles = generator.GenerateForLocation([]policies.Policy{test.policy}, http.Location{})
			g.Expect(resFiles).To(HaveLen(1))

			for _, str := range test.expStrings {
				g.Expect(string(resFiles[0].Content)).To(ContainSubstring(str))
			}

			resFiles = generator.GenerateForInternalLocation([]policies.Policy{test.policy})
			g.Expect(resFiles).To(HaveLen(1))

			for _, str := range test.expStrings {
				g.Expect(string(resFiles[0].Content)).To(ContainSubstring(str))
			}
		})
	}
}

func TestGenerateNoPolicies(t *testing.T) {
	g := NewWithT(t)

	generator := clientsettings.NewGenerator()

	resFiles := generator.GenerateForServer([]policies.Policy{}, http.Server{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForServer([]policies.Policy{&ngfAPI.ObservabilityPolicy{}}, http.Server{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForLocation([]policies.Policy{}, http.Location{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForLocation([]policies.Policy{&ngfAPI.ObservabilityPolicy{}}, http.Location{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForInternalLocation([]policies.Policy{})
	g.Expect(resFiles).To(BeEmpty())

	resFiles = generator.GenerateForInternalLocation([]policies.Policy{&ngfAPI.ObservabilityPolicy{}})
	g.Expect(resFiles).To(BeEmpty())
}
