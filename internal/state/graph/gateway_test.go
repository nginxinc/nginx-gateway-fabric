//nolint:gosec
package graph

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets/secretsfakes"
)

func TestProcessedGatewaysGetAllNsNames(t *testing.T) {
	winner := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-1",
		},
	}
	loser := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-2",
		},
	}

	tests := []struct {
		gws      processedGateways
		name     string
		expected []types.NamespacedName
	}{
		{
			gws:      processedGateways{},
			expected: nil,
			name:     "no gateways",
		},
		{
			gws: processedGateways{
				Winner: winner,
				Ignored: map[types.NamespacedName]*v1beta1.Gateway{
					client.ObjectKeyFromObject(loser): loser,
				},
			},
			expected: []types.NamespacedName{
				client.ObjectKeyFromObject(winner),
				client.ObjectKeyFromObject(loser),
			},
			name: "winner and ignored",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := test.gws.GetAllNsNames()
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestProcessGateways(t *testing.T) {
	const gcName = "test-gc"

	winner := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-1",
		},
		Spec: v1beta1.GatewaySpec{
			GatewayClassName: gcName,
		},
	}
	loser := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway-2",
		},
		Spec: v1beta1.GatewaySpec{
			GatewayClassName: gcName,
		},
	}

	tests := []struct {
		gws      map[types.NamespacedName]*v1beta1.Gateway
		expected processedGateways
		name     string
	}{
		{
			gws:      nil,
			expected: processedGateways{},
			name:     "no gateways",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "some-gateway"}: {
					Spec: v1beta1.GatewaySpec{GatewayClassName: "some-class"},
				},
			},
			expected: processedGateways{},
			name:     "unrelated gateway",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway-1"}: winner,
			},
			expected: processedGateways{
				Winner:  winner,
				Ignored: map[types.NamespacedName]*v1beta1.Gateway{},
			},
			name: "one gateway",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway-1"}: winner,
				{Namespace: "test", Name: "gateway-2"}: loser,
			},
			expected: processedGateways{
				Winner: winner,
				Ignored: map[types.NamespacedName]*v1beta1.Gateway{
					{Namespace: "test", Name: "gateway-2"}: loser,
				},
			},
			name: "multiple gateways",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := processGateways(test.gws, gcName)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestBuildGateway(t *testing.T) {
	const gcName = "my-gateway-class"

	labelSet := map[string]string{
		"key": "value",
	}
	listenerAllowedRoutes := v1beta1.Listener{
		Name:     "listener-with-allowed-routes",
		Hostname: (*v1beta1.Hostname)(helpers.GetPointer("foo.example.com")),
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
		AllowedRoutes: &v1beta1.AllowedRoutes{
			Kinds: []v1beta1.RouteGroupKind{
				{Kind: "HTTPRoute", Group: helpers.GetPointer(v1beta1.Group(v1beta1.GroupName))},
			},
			Namespaces: &v1beta1.RouteNamespaces{
				From:     helpers.GetPointer(v1beta1.NamespacesFromSelector),
				Selector: &metav1.LabelSelector{MatchLabels: labelSet},
			},
		},
	}
	listenerInvalidSelector := *listenerAllowedRoutes.DeepCopy()
	listenerInvalidSelector.Name = "listener-with-invalid-selector"
	listenerInvalidSelector.AllowedRoutes.Namespaces.Selector.MatchExpressions = []metav1.LabelSelectorRequirement{
		{
			Operator: "invalid",
		},
	}

	gatewayTLSConfigSameNs := &v1beta1.GatewayTLSConfig{
		Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
		CertificateRefs: []v1beta1.SecretObjectReference{
			{
				Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
				Name:      "secret",
				Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
			},
		},
	}

	tlsConfigInvalidSecret := &v1beta1.GatewayTLSConfig{
		Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
		CertificateRefs: []v1beta1.SecretObjectReference{
			{
				Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
				Name:      "does-not-exist",
				Namespace: (*v1beta1.Namespace)(helpers.GetPointer("test")),
			},
		},
	}

	gatewayTLSConfigDiffNs := &v1beta1.GatewayTLSConfig{
		Mode: helpers.GetPointer(v1beta1.TLSModeTerminate),
		CertificateRefs: []v1beta1.SecretObjectReference{
			{
				Kind:      (*v1beta1.Kind)(helpers.GetPointer("Secret")),
				Name:      "secret",
				Namespace: (*v1beta1.Namespace)(helpers.GetPointer("diff-ns")),
			},
		},
	}

	createListener := func(
		name string,
		hostname string,
		port int,
		protocol v1beta1.ProtocolType,
		tls *v1beta1.GatewayTLSConfig,
	) v1beta1.Listener {
		return v1beta1.Listener{
			Name:     v1beta1.SectionName(name),
			Hostname: (*v1beta1.Hostname)(helpers.GetPointer(hostname)),
			Port:     v1beta1.PortNumber(port),
			Protocol: protocol,
			TLS:      tls,
		}
	}
	createHTTPListener := func(name, hostname string, port int) v1beta1.Listener {
		return createListener(name, hostname, port, v1beta1.HTTPProtocolType, nil)
	}
	createTCPListener := func(name, hostname string, port int) v1beta1.Listener {
		return createListener(name, hostname, port, v1beta1.TCPProtocolType, nil)
	}
	createHTTPSListener := func(name, hostname string, port int, tls *v1beta1.GatewayTLSConfig) v1beta1.Listener {
		return createListener(name, hostname, port, v1beta1.HTTPSProtocolType, tls)
	}

	// foo http listeners
	foo80Listener1 := createHTTPListener("foo-80-1", "foo.example.com", 80)
	foo8080Listener := createHTTPListener("foo-8080", "foo.example.com", 8080)
	foo8081Listener := createHTTPListener("foo-8081", "foo.example.com", 8081)
	foo443Listener := createHTTPListener("foo-443", "foo.example.com", 443)

	// foo https listeners
	foo80HTTPSListener := createHTTPSListener("foo-80-https", "foo.example.com", 80, gatewayTLSConfigSameNs)
	foo443HTTPSListener1 := createHTTPSListener("foo-443-https-1", "foo.example.com", 443, gatewayTLSConfigSameNs)
	foo8443HTTPSListener := createHTTPSListener("foo-8443-https", "foo.example.com", 8443, gatewayTLSConfigSameNs)

	// bar http listener
	bar80Listener := createHTTPListener("bar-80", "bar.example.com", 80)

	// bar https listeners
	bar443HTTPSListener := createHTTPSListener("bar-443-https", "bar.example.com", 443, gatewayTLSConfigSameNs)
	bar8443HTTPSListener := createHTTPSListener("bar-8443-https", "bar.example.com", 8443, gatewayTLSConfigSameNs)

	// https listener that references secret in different namespace
	crossNamespaceSecretListener := createHTTPSListener(
		"listener-cross-ns-secret",
		"foo.example.com",
		443,
		gatewayTLSConfigDiffNs,
	)

	// invalid listeners
	invalidProtocolListener := createTCPListener("invalid-protocol", "bar.example.com", 80)
	invalidPortListener := createHTTPListener("invalid-port", "invalid-port", 0)
	invalidHostnameListener := createHTTPListener("invalid-hostname", "$example.com", 80)
	invalidHTTPSHostnameListener := createHTTPSListener(
		"invalid-https-hostname",
		"$example.com",
		443,
		gatewayTLSConfigSameNs,
	)
	invalidTLSConfigListener := createHTTPSListener(
		"invalid-tls-config",
		"foo.example.com",
		443,
		tlsConfigInvalidSecret,
	)
	invalidHTTPSPortListener := createHTTPSListener("invalid-https-port", "foo.example.com", 0, gatewayTLSConfigSameNs)

	const (
		invalidHostnameMsg = `hostname: Invalid value: "$example.com": a lowercase RFC 1123 subdomain ` +
			"must consist of lower case alphanumeric characters, '-' or '.', and must start and end " +
			"with an alphanumeric character (e.g. 'example.com', regex used for validation is " +
			`'[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`

		conflict80PortMsg = "Multiple listeners for the same port 80 specify incompatible protocols; " +
			"ensure only one protocol per port"

		conflict443PortMsg = "Multiple listeners for the same port 443 specify incompatible protocols; " +
			"ensure only one protocol per port"

		secretPath       = "/etc/nginx/secrets/test_secret"
		diffNsSecretPath = "/etc/nginx/secrets/diff_ns_secret"
	)

	type gatewayCfg struct {
		listeners []v1beta1.Listener
		addresses []v1beta1.GatewayAddress
	}

	var lastCreatedGateway *v1beta1.Gateway
	createGateway := func(cfg gatewayCfg) *v1beta1.Gateway {
		lastCreatedGateway = &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
			},
			Spec: v1beta1.GatewaySpec{
				GatewayClassName: gcName,
				Listeners:        cfg.listeners,
				Addresses:        cfg.addresses,
			},
		}
		return lastCreatedGateway
	}
	getLastCreatedGetaway := func() *v1beta1.Gateway {
		return lastCreatedGateway
	}

	validGC := &GatewayClass{
		Valid: true,
	}
	invalidGC := &GatewayClass{
		Valid: false,
	}

	tests := []struct {
		gateway      *v1beta1.Gateway
		gatewayClass *GatewayClass
		refGrants    map[types.NamespacedName]*v1beta1.ReferenceGrant
		expected     *Gateway
		name         string
	}{
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{foo80Listener1, foo8080Listener}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"foo-80-1": {
						Source: foo80Listener1,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
					"foo-8080": {
						Source: foo8080Listener,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
				},
				Valid: true,
			},
			name: "valid http listeners",
		},
		{
			gateway: createGateway(
				gatewayCfg{listeners: []v1beta1.Listener{foo443HTTPSListener1, foo8443HTTPSListener}},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"foo-443-https-1": {
						Source:     foo443HTTPSListener1,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
					"foo-8443-https": {
						Source:     foo8443HTTPSListener,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
				},
				Valid: true,
			},
			name: "valid https listeners",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{listenerAllowedRoutes}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"listener-with-allowed-routes": {
						Source:                    listenerAllowedRoutes,
						Valid:                     true,
						AllowedRouteLabelSelector: labels.SelectorFromSet(labels.Set(labelSet)),
						Routes:                    map[types.NamespacedName]*Route{},
					},
				},
				Valid: true,
			},
			name: "valid http listener with allowed routes label selector",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{crossNamespaceSecretListener}}),
			gatewayClass: validGC,
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Name: "ref-grant", Namespace: "diff-ns"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ref-grant",
						Namespace: "diff-ns",
					},
					Spec: v1beta1.ReferenceGrantSpec{
						From: []v1beta1.ReferenceGrantFrom{
							{
								Group:     v1beta1.GroupName,
								Kind:      "Gateway",
								Namespace: "test",
							},
						},
						To: []v1beta1.ReferenceGrantTo{
							{
								Group: "core",
								Kind:  "Secret",
								Name:  helpers.GetPointer(v1beta1.ObjectName("secret")),
							},
						},
					},
				},
			},
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"listener-cross-ns-secret": {
						Source:     crossNamespaceSecretListener,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: diffNsSecretPath,
					},
				},
				Valid: true,
			},
			name: "valid https listener with cross-namespace secret; allowed by reference grant",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{crossNamespaceSecretListener}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"listener-cross-ns-secret": {
						Source: crossNamespaceSecretListener,
						Valid:  false,
						Conditions: conditions.NewListenerRefNotPermitted(
							`Certificate ref to secret diff-ns/secret not permitted by any ReferenceGrant`,
						),
						Routes: map[types.NamespacedName]*Route{},
					},
				},
				Valid: true,
			},
			name: "invalid https listener with cross-namespace secret; no reference grant",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{listenerInvalidSelector}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"listener-with-invalid-selector": {
						Source: listenerInvalidSelector,
						Valid:  false,
						Conditions: conditions.NewListenerUnsupportedValue(
							`invalid label selector: "invalid" is not a valid label selector operator`,
						),
					},
				},
				Valid: true,
			},
			name: "http listener with invalid label selector",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{invalidProtocolListener}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"invalid-protocol": {
						Source: invalidProtocolListener,
						Valid:  false,
						Conditions: conditions.NewListenerUnsupportedProtocol(
							`protocol: Unsupported value: "TCP": supported values: "HTTP", "HTTPS"`,
						),
					},
				},
				Valid: true,
			},
			name: "invalid listener protocol",
		},
		{
			gateway: createGateway(
				gatewayCfg{listeners: []v1beta1.Listener{invalidPortListener, invalidHTTPSPortListener}},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"invalid-port": {
						Source: invalidPortListener,
						Valid:  false,
						Conditions: conditions.NewListenerUnsupportedValue(
							`port: Invalid value: 0: port must be between 1-65535`,
						),
					},
					"invalid-https-port": {
						Source: invalidHTTPSPortListener,
						Valid:  false,
						Conditions: conditions.NewListenerUnsupportedValue(
							`port: Invalid value: 0: port must be between 1-65535`,
						),
					},
				},
				Valid: true,
			},
			name: "invalid ports",
		},
		{
			gateway: createGateway(
				gatewayCfg{listeners: []v1beta1.Listener{invalidHostnameListener, invalidHTTPSHostnameListener}},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"invalid-hostname": {
						Source:     invalidHostnameListener,
						Valid:      false,
						Conditions: conditions.NewListenerUnsupportedValue(invalidHostnameMsg),
					},
					"invalid-https-hostname": {
						Source:     invalidHTTPSHostnameListener,
						Valid:      false,
						Conditions: conditions.NewListenerUnsupportedValue(invalidHostnameMsg),
					},
				},
				Valid: true,
			},
			name: "invalid hostnames",
		},
		{
			gateway:      createGateway(gatewayCfg{listeners: []v1beta1.Listener{invalidTLSConfigListener}}),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"invalid-tls-config": {
						Source: invalidTLSConfigListener,
						Valid:  false,
						Routes: map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerInvalidCertificateRef(
							`tls.certificateRefs[0]: Invalid value: test/does-not-exist: secret not found`,
						),
					},
				},
				Valid: true,
			},
			name: "invalid https listener (secret does not exist)",
		},
		{
			gateway: createGateway(
				gatewayCfg{
					listeners: []v1beta1.Listener{
						foo80Listener1,
						foo8080Listener,
						foo8081Listener,
						foo443HTTPSListener1,
						foo8443HTTPSListener,
						bar80Listener,
						bar443HTTPSListener,
						bar8443HTTPSListener,
					},
				},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"foo-80-1": {
						Source: foo80Listener1,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
					"foo-8080": {
						Source: foo8080Listener,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
					"foo-8081": {
						Source: foo8081Listener,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
					"bar-80": {
						Source: bar80Listener,
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{},
					},
					"foo-443-https-1": {
						Source:     foo443HTTPSListener1,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
					"foo-8443-https": {
						Source:     foo8443HTTPSListener,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
					"bar-443-https": {
						Source:     bar443HTTPSListener,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
					"bar-8443-https": {
						Source:     bar8443HTTPSListener,
						Valid:      true,
						Routes:     map[types.NamespacedName]*Route{},
						SecretPath: secretPath,
					},
				},
				Valid: true,
			},
			name: "multiple valid http/https listeners",
		},
		{
			gateway: createGateway(
				gatewayCfg{
					listeners: []v1beta1.Listener{
						foo80Listener1,
						bar80Listener,
						foo443Listener,
						foo80HTTPSListener,
						foo443HTTPSListener1,
						bar443HTTPSListener,
					},
				},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Listeners: map[string]*Listener{
					"foo-80-1": {
						Source:     foo80Listener1,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict80PortMsg),
					},
					"bar-80": {
						Source:     bar80Listener,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict80PortMsg),
					},
					"foo-443": {
						Source:     foo443Listener,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict443PortMsg),
					},
					"foo-80-https": {
						Source:     foo80HTTPSListener,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict80PortMsg),
						SecretPath: "/etc/nginx/secrets/test_secret",
					},
					"foo-443-https-1": {
						Source:     foo443HTTPSListener1,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict443PortMsg),
						SecretPath: "/etc/nginx/secrets/test_secret",
					},
					"bar-443-https": {
						Source:     bar443HTTPSListener,
						Valid:      false,
						Routes:     map[types.NamespacedName]*Route{},
						Conditions: conditions.NewListenerProtocolConflict(conflict443PortMsg),
						SecretPath: "/etc/nginx/secrets/test_secret",
					},
				},
				Valid: true,
			},
			name: "port/protocol collisions",
		},
		{
			gateway: createGateway(
				gatewayCfg{
					listeners: []v1beta1.Listener{foo80Listener1, foo443HTTPSListener1},
					addresses: []v1beta1.GatewayAddress{{}},
				},
			),
			gatewayClass: validGC,
			expected: &Gateway{
				Source: getLastCreatedGetaway(),
				Valid:  false,
				Conditions: conditions.NewGatewayUnsupportedValue("spec." +
					"addresses: Forbidden: addresses are not supported",
				),
			},
			name: "gateway addresses are not supported",
		},
		{
			gateway:  nil,
			expected: nil,
			name:     "nil gateway",
		},
		{
			gateway: createGateway(
				gatewayCfg{listeners: []v1beta1.Listener{foo80Listener1, invalidProtocolListener}},
			),
			gatewayClass: invalidGC,
			expected: &Gateway{
				Source:     getLastCreatedGetaway(),
				Valid:      false,
				Conditions: conditions.NewGatewayInvalid("GatewayClass is invalid"),
			},
			name: "invalid gatewayclass",
		},
		{
			gateway: createGateway(
				gatewayCfg{listeners: []v1beta1.Listener{foo80Listener1, invalidProtocolListener}},
			),
			gatewayClass: nil,
			expected: &Gateway{
				Source:     getLastCreatedGetaway(),
				Valid:      false,
				Conditions: conditions.NewGatewayInvalid("GatewayClass doesn't exist"),
			},
			name: "nil gatewayclass",
		},
	}

	secretMemoryMgr := &secretsfakes.FakeSecretDiskMemoryManager{}
	secretMemoryMgr.RequestCalls(func(nsname types.NamespacedName) (string, error) {
		switch nsname {
		case types.NamespacedName{Namespace: "test", Name: "secret"}:
			return secretPath, nil
		case types.NamespacedName{Namespace: "diff-ns", Name: "secret"}:
			return diffNsSecretPath, nil
		default:
			return "", errors.New("secret not found")
		}
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := buildGateway(test.gateway, secretMemoryMgr, test.gatewayClass, test.refGrants)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
