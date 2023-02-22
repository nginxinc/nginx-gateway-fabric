package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

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
		gws                map[types.NamespacedName]*v1beta1.Gateway
		expectedWinner     *v1beta1.Gateway
		expectedIgnoredGws map[types.NamespacedName]*v1beta1.Gateway
		msg                string
	}{
		{
			gws:                nil,
			expectedWinner:     nil,
			expectedIgnoredGws: nil,
			msg:                "no gateways",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "some-gateway"}: {
					Spec: v1beta1.GatewaySpec{GatewayClassName: "some-class"},
				},
			},
			expectedWinner:     nil,
			expectedIgnoredGws: nil,
			msg:                "unrelated gateway",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway"}: winner,
			},
			expectedWinner:     winner,
			expectedIgnoredGws: map[types.NamespacedName]*v1beta1.Gateway{},
			msg:                "one gateway",
		},
		{
			gws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway-1"}: winner,
				{Namespace: "test", Name: "gateway-2"}: loser,
			},
			expectedWinner: winner,
			expectedIgnoredGws: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway-2"}: loser,
			},
			msg: "multiple gateways",
		},
	}

	for _, test := range tests {
		winner, ignoredGws := processGateways(test.gws, gcName)

		if diff := cmp.Diff(winner, test.expectedWinner); diff != "" {
			t.Errorf("processGateways() '%s' mismatch for winner (-want +got):\n%s", test.msg, diff)
		}
		if diff := cmp.Diff(ignoredGws, test.expectedIgnoredGws); diff != "" {
			t.Errorf("processGateways() '%s' mismatch for ignored gateways (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestBuildListeners(t *testing.T) {
	const gcName = "my-gateway-class"

	listener801 := v1beta1.Listener{
		Name:     "listener-80-1",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}
	listener802 := v1beta1.Listener{
		Name:     "listener-80-2",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1beta1.TCPProtocolType, // invalid protocol
	}
	listener803 := v1beta1.Listener{
		Name:     "listener-80-3",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}
	listener804 := v1beta1.Listener{
		Name:     "listener-80-4",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}
	listener805 := v1beta1.Listener{
		Name:     "listener-80-5",
		Port:     81, // invalid port
		Protocol: v1beta1.HTTPProtocolType,
	}
	listener806 := v1beta1.Listener{
		Name:     "listener-80-6",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("$example.com")), // invalid hostname
		Port:     80,
		Protocol: v1beta1.HTTPProtocolType,
	}

	gatewayTLSConfig := &v1beta1.GatewayTLSConfig{
		Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
		CertificateRefs: []v1beta1.SecretObjectReference{
			{
				Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
				Name:      "secret",
				Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
			},
		},
	}

	tlsConfigInvalidSecret := &v1beta1.GatewayTLSConfig{
		Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
		CertificateRefs: []v1beta1.SecretObjectReference{
			{
				Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
				Name:      "does-not-exist",
				Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
			},
		},
	}
	// https listeners
	listener4431 := v1beta1.Listener{
		Name:     "listener-443-1",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1beta1.HTTPSProtocolType,
	}
	listener4432 := v1beta1.Listener{
		Name:     "listener-443-2",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("bar.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1beta1.HTTPSProtocolType,
	}
	listener4433 := v1beta1.Listener{
		Name:     "listener-443-3",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1beta1.HTTPSProtocolType,
	}
	listener4434 := v1beta1.Listener{
		Name:     "listener-443-4",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("$example.com")), // invalid hostname
		Port:     443,
		TLS:      gatewayTLSConfig,
		Protocol: v1beta1.HTTPSProtocolType,
	}
	listener4435 := v1beta1.Listener{
		Name:     "listener-443-5",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     443,
		TLS:      tlsConfigInvalidSecret, // invalid https listener; secret does not exist
		Protocol: v1beta1.HTTPSProtocolType,
	}
	listener4436 := v1beta1.Listener{
		Name:     "listener-443-6",
		Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
		Port:     444, // invalid port
		TLS:      gatewayTLSConfig,
		Protocol: v1beta1.HTTPSProtocolType,
	}

	const (
		invalidHostnameMsg = "Invalid hostname: a lowercase RFC 1123 subdomain " +
			"must consist of lower case alphanumeric characters, '-' or '.', and must start and end " +
			"with an alphanumeric character (e.g. 'example.com', regex used for validation is " +
			`'[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')`

		conflictedHostnamesMsg = `Multiple listeners for the same port use the same hostname "foo.example.com"; ` +
			"ensure only one listener uses that hostname"
	)

	tests := []struct {
		gateway  *v1beta1.Gateway
		expected map[string]*Listener
		name     string
	}{
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener801,
					},
				},
				Status: v1beta1.GatewayStatus{},
			},
			expected: map[string]*Listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
				},
			},
			name: "valid http listener",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener4431,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-443-1": {
					Source:            listener4431,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
			},
			name: "valid https listener",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener802,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-2": {
					Source:            listener802,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerUnsupportedProtocol(`Protocol "TCP" is not supported, use "HTTP" ` +
							`or "HTTPS"`),
					},
				},
			},
			name: "invalid listener protocol",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener805,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-5": {
					Source:            listener805,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerPortUnavailable("Port 81 is not supported for HTTP, use 80"),
					},
				},
			},
			name: "invalid http listener",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener4436,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-443-6": {
					Source:            listener4436,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerPortUnavailable("Port 444 is not supported for HTTPS, use 443"),
					},
				},
			},
			name: "invalid https listener",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener806,
						listener4434,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-6": {
					Source:            listener806,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerUnsupportedValue(invalidHostnameMsg),
					},
				},
				"listener-443-4": {
					Source:            listener4434,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerUnsupportedValue(invalidHostnameMsg),
					},
				},
			},
			name: "invalid hostnames",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener4435,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-443-5": {
					Source:            listener4435,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: conditions.NewListenerInvalidCertificateRef("Failed to get the certificate " +
						"test/does-not-exist: secret test/does-not-exist does not exist"),
				},
			},
			name: "invalid https listener (secret does not exist)",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener801, listener803,
						listener4431, listener4432,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-80-3": {
					Source:            listener803,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
				},
				"listener-443-1": {
					Source:            listener4431,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
				"listener-443-2": {
					Source:            listener4432,
					Valid:             true,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        secretPath,
				},
			},
			name: "multiple valid http/https listeners",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener801, listener804,
						listener4431, listener4433,
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions:        conditions.NewListenerConflictedHostname(conflictedHostnamesMsg),
				},
				"listener-80-4": {
					Source:            listener804,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions:        conditions.NewListenerConflictedHostname(conflictedHostnamesMsg),
				},
				"listener-443-1": {
					Source:            listener4431,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions:        conditions.NewListenerConflictedHostname(conflictedHostnamesMsg),
				},
				"listener-443-3": {
					Source:            listener4433,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions:        conditions.NewListenerConflictedHostname(conflictedHostnamesMsg),
				},
			},
			name: "collisions",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: gcName,
					Listeners: []v1beta1.Listener{
						listener801,
						listener4431,
					},
					Addresses: []v1beta1.GatewayAddress{
						{},
					},
				},
			},
			expected: map[string]*Listener{
				"listener-80-1": {
					Source:            listener801,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					Conditions: []conditions.Condition{
						conditions.NewListenerUnsupportedAddress("Specifying Gateway addresses is not supported"),
					},
				},
				"listener-443-1": {
					Source:            listener4431,
					Valid:             false,
					Routes:            map[types.NamespacedName]*Route{},
					AcceptedHostnames: map[string]struct{}{},
					SecretPath:        "",
					Conditions: []conditions.Condition{
						conditions.NewListenerUnsupportedAddress("Specifying Gateway addresses is not supported"),
					},
				},
			},
			name: "gateway addresses are not supported",
		},
		{
			gateway:  nil,
			expected: map[string]*Listener{},
			name:     "no gateway",
		},
		{
			gateway: &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
				},
				Spec: v1beta1.GatewaySpec{
					GatewayClassName: "wrong-class",
					Listeners: []v1beta1.Listener{
						listener801, listener804,
					},
				},
			},
			expected: map[string]*Listener{},
			name:     "wrong gatewayclass",
		},
	}

	// add secret to store
	secretStore := secrets.NewSecretStore()
	secretStore.Upsert(testSecret)

	secretMemoryMgr := secrets.NewSecretDiskMemoryManager(secretsDirectory, secretStore)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := buildListeners(test.gateway, gcName, secretMemoryMgr)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestValidateHTTPListener(t *testing.T) {
	tests := []struct {
		l        v1beta1.Listener
		name     string
		expected []conditions.Condition
	}{
		{
			l: v1beta1.Listener{
				Port: 80,
			},
			expected: nil,
			name:     "valid",
		},
		{
			l: v1beta1.Listener{
				Port: 81,
			},
			expected: []conditions.Condition{
				conditions.NewListenerPortUnavailable("Port 81 is not supported for HTTP, use 80"),
			},
			name: "invalid port",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := validateHTTPListener(test.l)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateHTTPSListener(t *testing.T) {
	gwNs := "gateway-ns"

	validSecretRef := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefGroup := v1beta1.SecretObjectReference{
		Group:     (*v1beta1.Group)(helpers.GetStringPointer("some-group")),
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefKind := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("ConfigMap")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer(gwNs)),
	}

	invalidSecretRefTNamespace := v1beta1.SecretObjectReference{
		Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
		Name:      "secret",
		Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("diff-ns")),
	}

	tests := []struct {
		l        v1beta1.Listener
		name     string
		expected []conditions.Condition
	}{
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: nil,
			name:     "valid",
		},
		{
			l: v1beta1.Listener{
				Port: 80,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerPortUnavailable("Port 80 is not supported for HTTPS, use 443"),
			},
			name: "invalid port",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
					Options:         map[v1beta1.AnnotationKey]v1beta1.AnnotationValue{"key": "val"},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue("tls.options are not supported"),
			},
			name: "invalid options",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModePassthrough),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue(`tls.mode "Passthrough" is not supported, use "Terminate"`),
			},
			name: "invalid tls mode",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefGroup},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(`Group must be empty, got "some-group"`),
			name:     "invalid cert ref group",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefKind},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef(`Kind must be Secret, got "ConfigMap"`),
			name:     "invalid cert ref kind",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{invalidSecretRefTNamespace},
				},
			},
			expected: conditions.NewListenerInvalidCertificateRef("Referenced Secret must belong to the same " +
				"namespace as the Gateway"),
			name: "invalid cert ref namespace",
		},
		{
			l: v1beta1.Listener{
				Port: 443,
				TLS: &v1beta1.GatewayTLSConfig{
					Mode:            helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
					CertificateRefs: []v1beta1.SecretObjectReference{validSecretRef, validSecretRef},
				},
			},
			expected: []conditions.Condition{
				conditions.NewListenerUnsupportedValue("Only 1 certificateRef is supported, got 2"),
			},
			name: "too many cert refs",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := validateHTTPSListener(test.l, gwNs)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateListenerHostname(t *testing.T) {
	tests := []struct {
		hostname  *v1beta1.Hostname
		name      string
		expectErr bool
	}{
		{
			hostname:  nil,
			expectErr: false,
			name:      "nil hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("")),
			expectErr: false,
			name:      "empty hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			expectErr: false,
			name:      "valid hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("*.example.com")),
			expectErr: true,
			name:      "wildcard hostname",
		},
		{
			hostname:  (*v1beta1.Hostname)(helpers.GetStringPointer("example$com")),
			expectErr: true,
			name:      "invalid hostname",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateListenerHostname(test.hostname)

			if test.expectErr {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}
