//nolint:gosec
package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

var testSecret = &v1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "secret",
		Namespace: "test",
	},
	Data: map[string][]byte{
		v1.TLSCertKey: []byte(`-----BEGIN CERTIFICATE-----
MIIDLjCCAhYCCQDAOF9tLsaXWjANBgkqhkiG9w0BAQsFADBaMQswCQYDVQQGEwJV
UzELMAkGA1UECAwCQ0ExITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0
ZDEbMBkGA1UEAwwSY2FmZS5leGFtcGxlLmNvbSAgMB4XDTE4MDkxMjE2MTUzNVoX
DTIzMDkxMTE2MTUzNVowWDELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAkNBMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQxGTAXBgNVBAMMEGNhZmUuZXhh
bXBsZS5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCp6Kn7sy81
p0juJ/cyk+vCAmlsfjtFM2muZNK0KtecqG2fjWQb55xQ1YFA2XOSwHAYvSdwI2jZ
ruW8qXXCL2rb4CZCFxwpVECrcxdjm3teViRXVsYImmJHPPSyQgpiobs9x7DlLc6I
BA0ZjUOyl0PqG9SJexMV73WIIa5rDVSF2r4kSkbAj4Dcj7LXeFlVXH2I5XwXCptC
n67JCg42f+k8wgzcRVp8XZkZWZVjwq9RUKDXmFB2YyN1XEWdZ0ewRuKYUJlsm692
skOrKQj0vkoPn41EE/+TaVEpqLTRoUY3rzg7DkdzfdBizFO2dsPNFx2CW0jXkNLv
Ko25CZrOhXAHAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAKHFCcyOjZvoHswUBMdL
RdHIb383pWFynZq/LuUovsVA58B0Cg7BEfy5vWVVrq5RIkv4lZ81N29x21d1JH6r
jSnQx+DXCO/TJEV5lSCUpIGzEUYaUPgRyjsM/NUdCJ8uHVhZJ+S6FA+CnOD9rn2i
ZBePCI5rHwEXwnnl8ywij3vvQ5zHIuyBglWr/Qyui9fjPpwWUvUm4nv5SMG9zCV7
PpuwvuatqjO1208BjfE/cZHIg8Hw9mvW9x9C+IQMIMDE7b/g6OcK7LGTLwlFxvA8
7WjEequnayIphMhKRXVf1N349eN98Ez38fOTHTPbdJjFA/PcC+Gyme+iGt5OQdFh
yRE=
-----END CERTIFICATE-----`),
		v1.TLSPrivateKeyKey: []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAqeip+7MvNadI7if3MpPrwgJpbH47RTNprmTStCrXnKhtn41k
G+ecUNWBQNlzksBwGL0ncCNo2a7lvKl1wi9q2+AmQhccKVRAq3MXY5t7XlYkV1bG
CJpiRzz0skIKYqG7Pcew5S3OiAQNGY1DspdD6hvUiXsTFe91iCGuaw1Uhdq+JEpG
wI+A3I+y13hZVVx9iOV8FwqbQp+uyQoONn/pPMIM3EVafF2ZGVmVY8KvUVCg15hQ
dmMjdVxFnWdHsEbimFCZbJuvdrJDqykI9L5KD5+NRBP/k2lRKai00aFGN684Ow5H
c33QYsxTtnbDzRcdgltI15DS7yqNuQmazoVwBwIDAQABAoIBAQCPSdSYnQtSPyql
FfVFpTOsoOYRhf8sI+ibFxIOuRauWehhJxdm5RORpAzmCLyL5VhjtJme223gLrw2
N99EjUKb/VOmZuDsBc6oCF6QNR58dz8cnORTewcotsJR1pn1hhlnR5HqJJBJask1
ZEnUQfcXZrL94lo9JH3E+Uqjo1FFs8xxE8woPBqjZsV7pRUZgC3LhxnwLSExyFo4
cxb9SOG5OmAJozStFoQ2GJOes8rJ5qfdvytgg9xbLaQL/x0kpQ62BoFMBDdqOePW
KfP5zZ6/07/vpj48yA1Q32PzobubsBLd3Kcn32jfm1E7prtWl+JeOFiOznBQFJbN
4qPVRz5hAoGBANtWyxhNCSLu4P+XgKyckljJ6F5668fNj5CzgFRqJ09zn0TlsNro
FTLZcxDqnR3HPYM42JERh2J/qDFZynRQo3cg3oeivUdBVGY8+FI1W0qdub/L9+yu
edOZTQ5XmGGp6r6jexymcJim/OsB3ZnYOpOrlD7SPmBvzNLk4MF6gxbXAoGBAMZO
0p6HbBmcP0tjFXfcKE77ImLm0sAG4uHoUx0ePj/2qrnTnOBBNE4MvgDuTJzy+caU
k8RqmdHCbHzTe6fzYq/9it8sZ77KVN1qkbIcuc+RTxA9nNh1TjsRne74Z0j1FCLk
hHcqH0ri7PYSKHTE8FvFCxZYdbuB84CmZihvxbpRAoGAIbjqaMYPTYuklCda5S79
YSFJ1JzZe1Kja//tDw1zFcgVCKa31jAwciz0f/lSRq3HS1GGGmezhPVTiqLfeZqc
R0iKbhgbOcVVkJJ3K0yAyKwPTumxKHZ6zImZS0c0am+RY9YGq5T7YrzpzcfvpiOU
ffe3RyFT7cfCmfoOhDCtzukCgYB30oLC1RLFOrqn43vCS51zc5zoY44uBzspwwYN
TwvP/ExWMf3VJrDjBCH+T/6sysePbJEImlzM+IwytFpANfiIXEt/48Xf60Nx8gWM
uHyxZZx/NKtDw0V8vX1POnq2A5eiKa+8jRARYKJLYNdfDuwolxvG6bZhkPi/4EtT
3Y18sQKBgHtKbk+7lNJVeswXE5cUG6EDUsDe/2Ua7fXp7FcjqBEoap1LSw+6TXp0
ZgrmKE8ARzM47+EJHUviiq/nupE15g0kJW3syhpU9zZLO7ltB0KIkO9ZRcmUjo8Q
cpLlHMAqbLJ8WYGJCkhiWxyal6hYTyWY4cVkC0xtTl/hUE9IeNKo
-----END RSA PRIVATE KEY-----`),
	},
	Type: v1.SecretTypeTLS,
}

var (
	secretPath       = "/etc/nginx/secrets/test_secret"
	secretsDirectory = "/etc/nginx/secrets"
)

func TestBuildGraph(t *testing.T) {
	const (
		gcName         = "my-class"
		controllerName = "my.controller"
	)

	createRoute := func(name string, gatewayName string, listenerName string) *v1beta1.HTTPRoute {
		return &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: []v1beta1.ParentReference{
						{
							Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
							Name:        v1beta1.ObjectName(gatewayName),
							SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer(listenerName)),
						},
					},
				},
				Hostnames: []v1beta1.Hostname{
					"foo.example.com",
				},
				Rules: []v1beta1.HTTPRouteRule{
					{
						Matches: []v1beta1.HTTPRouteMatch{
							{
								Path: &v1beta1.HTTPPathMatch{
									Value: helpers.GetStringPointer("/"),
								},
							},
						},
						BackendRefs: []v1beta1.HTTPBackendRef{
							{
								BackendRef: v1beta1.BackendRef{
									BackendObjectReference: v1beta1.BackendObjectReference{
										Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Service")),
										Name:      "foo",
										Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
										Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(80)),
									},
								},
							},
						},
					},
				},
			},
		}
	}

	hr1 := createRoute("hr-1", "gateway-1", "listener-80-1")
	hr2 := createRoute("hr-2", "wrong-gateway", "listener-80-1")
	hr3 := createRoute("hr-3", "gateway-1", "listener-443-1") // https listener; should not conflict with hr1

	fooSvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test"}}

	hr1Group := BackendGroup{
		Errors:  []string{},
		Source:  types.NamespacedName{Namespace: hr1.Namespace, Name: hr1.Name},
		RuleIdx: 0,
		Backends: []BackendRef{
			{
				Name:   "test_foo_80",
				Svc:    fooSvc,
				Port:   80,
				Valid:  true,
				Weight: 1,
			},
		},
	}

	hr3Group := BackendGroup{
		Errors:  []string{},
		Source:  types.NamespacedName{Namespace: hr3.Namespace, Name: hr3.Name},
		RuleIdx: 0,
		Backends: []BackendRef{
			{
				Name:   "test_foo_80",
				Svc:    fooSvc,
				Port:   80,
				Valid:  true,
				Weight: 1,
			},
		},
	}

	createGateway := func(name string) *v1beta1.Gateway {
		return &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1beta1.GatewaySpec{
				GatewayClassName: gcName,
				Listeners: []v1beta1.Listener{
					{
						Name:     "listener-80-1",
						Hostname: nil,
						Port:     80,
						Protocol: v1beta1.HTTPProtocolType,
					},

					{
						Name:     "listener-443-1",
						Hostname: nil,
						Port:     443,
						TLS: &v1beta1.GatewayTLSConfig{
							Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
							CertificateRefs: []v1beta1.SecretObjectReference{
								{
									Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
									Name:      "secret",
									Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
								},
							},
						},
						Protocol: v1beta1.HTTPSProtocolType,
					},
				},
			},
		}
	}

	gw1 := createGateway("gateway-1")
	gw2 := createGateway("gateway-2")

	svc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"}}

	store := ClusterStore{
		GatewayClass: &v1beta1.GatewayClass{
			Spec: v1beta1.GatewayClassSpec{
				ControllerName: controllerName,
			},
		},
		Gateways: map[types.NamespacedName]*v1beta1.Gateway{
			{Namespace: "test", Name: "gateway-1"}: gw1,
			{Namespace: "test", Name: "gateway-2"}: gw2,
		},
		HTTPRoutes: map[types.NamespacedName]*v1beta1.HTTPRoute{
			{Namespace: "test", Name: "hr-1"}: hr1,
			{Namespace: "test", Name: "hr-2"}: hr2,
			{Namespace: "test", Name: "hr-3"}: hr3,
		},
		Services: map[types.NamespacedName]*v1.Service{
			{Namespace: "test", Name: "foo"}: svc,
		},
	}

	routeHR1 := &Route{
		Source: hr1,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-80-1": {},
		},
		InvalidSectionNameRefs: map[string]conditions.Condition{},
		BackendGroups:          []BackendGroup{hr1Group},
	}

	routeHR3 := &Route{
		Source: hr3,
		ValidSectionNameRefs: map[string]struct{}{
			"listener-443-1": {},
		},
		InvalidSectionNameRefs: map[string]conditions.Condition{},
		BackendGroups:          []BackendGroup{hr3Group},
	}

	// add test secret to store
	secretStore := secrets.NewSecretStore()
	secretStore.Upsert(testSecret)
	secretRequestMgr := secrets.NewRequestManagerImpl(secretsDirectory, secretStore)

	expected := &Graph{
		GatewayClass: &GatewayClass{
			Source: store.GatewayClass,
			Valid:  true,
		},
		Gateway: &Gateway{
			Source: gw1,
			Listeners: map[string]*Listener{
				"listener-80-1": {
					Source: gw1.Spec.Listeners[0],
					Valid:  true,
					Routes: map[types.NamespacedName]*Route{
						{Namespace: "test", Name: "hr-1"}: routeHR1,
					},
					AcceptedHostnames: map[string]struct{}{
						"foo.example.com": {},
					},
				},
				"listener-443-1": {
					Source: gw1.Spec.Listeners[1],
					Valid:  true,
					Routes: map[types.NamespacedName]*Route{
						{Namespace: "test", Name: "hr-3"}: routeHR3,
					},
					AcceptedHostnames: map[string]struct{}{
						"foo.example.com": {},
					},
					SecretPath: secretPath,
				},
			},
		},
		IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
			{Namespace: "test", Name: "gateway-2"}: gw2,
		},
		Routes: map[types.NamespacedName]*Route{
			{Namespace: "test", Name: "hr-1"}: routeHR1,
			{Namespace: "test", Name: "hr-3"}: routeHR3,
		},
	}

	result := BuildGraph(store, controllerName, gcName, secretRequestMgr)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("BuildGraph() mismatch (-want +got):\n%s", diff)
	}
}
