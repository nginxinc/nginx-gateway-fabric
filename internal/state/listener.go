package state

import (
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

// listener represents a listener of the Gateway resource.
// FIXME(pleshakov) For now, we only support HTTP and HTTPS listeners.
type listener struct {
	// Source holds the source of the listener from the Gateway resource.
	Source v1beta1.Listener
	// Routes holds the routes attached to the listener.
	Routes map[types.NamespacedName]*route
	// AcceptedHostnames is an intersection between the hostnames supported by the listener and the hostnames
	// from the attached routes.
	AcceptedHostnames map[string]struct{}
	// SecretPath is the path to the secret on disk.
	SecretPath string
	// Conditions holds the conditions of the listener.
	Conditions []conditions.Condition
	// Valid shows whether the listener is valid.
	// A listener is considered valid if NKG can generate valid NGINX configuration for it.
	Valid bool
}

type listenerConfigurator interface {
	configure(listener v1beta1.Listener) *listener
}

type listenerConfiguratorFactory struct {
	https *httpListenerConfigurator
	http  *httpListenerConfigurator
}

func (f *listenerConfiguratorFactory) getConfiguratorForListener(l v1beta1.Listener) listenerConfigurator {
	switch l.Protocol {
	case v1beta1.HTTPProtocolType:
		return f.http
	case v1beta1.HTTPSProtocolType:
		return f.https
	default:
		return newInvalidProtocolListenerConfigurator()
	}
}

func newListenerConfiguratorFactory(
	gw *v1beta1.Gateway,
	secretMemoryMgr SecretDiskMemoryManager,
) *listenerConfiguratorFactory {
	return &listenerConfiguratorFactory{
		https: newHTTPSListenerConfigurator(gw, secretMemoryMgr),
		http:  newHTTPListenerConfigurator(gw),
	}
}

type httpListenerConfigurator struct {
	gateway         *v1beta1.Gateway
	secretMemoryMgr SecretDiskMemoryManager
	usedHostnames   map[string]*listener
	validate        func(gl v1beta1.Listener) []conditions.Condition
}

func newHTTPListenerConfigurator(gw *v1beta1.Gateway) *httpListenerConfigurator {
	return &httpListenerConfigurator{
		usedHostnames: make(map[string]*listener),
		gateway:       gw,
		validate:      validateHTTPListener,
	}
}

func newHTTPSListenerConfigurator(
	gateway *v1beta1.Gateway,
	secretMemoryMgr SecretDiskMemoryManager,
) *httpListenerConfigurator {
	return &httpListenerConfigurator{
		gateway:         gateway,
		secretMemoryMgr: secretMemoryMgr,
		usedHostnames:   make(map[string]*listener),
		validate: func(gl v1beta1.Listener) []conditions.Condition {
			return validateHTTPSListener(gl, gateway.Namespace)
		},
	}
}

func (c *httpListenerConfigurator) configure(gl v1beta1.Listener) *listener {
	conds := c.validate(gl)

	if len(c.gateway.Spec.Addresses) > 0 {
		conds = append(conds, conditions.NewListenerUnsupportedAddress("Specifying Gateway addresses is not supported"))
	}

	validHostnameErr := validateListenerHostname(gl.Hostname)
	if validHostnameErr != nil {
		msg := fmt.Sprintf("Invalid hostname: %v", validHostnameErr)
		conds = append(conds, conditions.NewListenerUnsupportedValue(msg))
	}

	l := &listener{
		Source:            gl,
		Valid:             len(conds) == 0,
		Routes:            make(map[types.NamespacedName]*route),
		AcceptedHostnames: make(map[string]struct{}),
		Conditions:        conds,
	}

	// this ensures that we don't check for hostname collisions for invalid hostnames
	if validHostnameErr != nil {
		return l
	}

	h := getHostname(gl.Hostname)
	if holder, exist := c.usedHostnames[h]; exist {
		l.Valid = false

		holder.Valid = false   // all listeners for the same hostname become conflicted
		holder.SecretPath = "" // ensure secret path is unset for invalid listeners

		format := "Multiple listeners for the same port use the same hostname %q; " +
			"ensure only one listener uses that hostname"
		conflictedConds := conditions.NewListenerConflictedHostname(fmt.Sprintf(format, h))

		holder.Conditions = append(holder.Conditions, conflictedConds...)
		l.Conditions = append(l.Conditions, conflictedConds...)
	} else {
		c.usedHostnames[h] = l
	}

	if !l.Valid {
		return l
	}

	if gl.Protocol == v1beta1.HTTPSProtocolType {
		nsname := types.NamespacedName{
			Namespace: c.gateway.Namespace,
			Name:      string(gl.TLS.CertificateRefs[0].Name),
		}

		var err error

		l.SecretPath, err = c.secretMemoryMgr.Request(nsname)
		if err != nil {
			msg := fmt.Sprintf("Failed to get the certificate %s: %v", nsname.String(), err)
			l.Conditions = append(l.Conditions, conditions.NewListenerInvalidCertificateRef(msg)...)
			l.Valid = false
		}
	}

	return l
}

type invalidProtocolListenerConfigurator struct{}

func newInvalidProtocolListenerConfigurator() *invalidProtocolListenerConfigurator {
	return &invalidProtocolListenerConfigurator{}
}

func (c *invalidProtocolListenerConfigurator) configure(gl v1beta1.Listener) *listener {
	msg := fmt.Sprintf("Protocol %q is not supported, use %q or %q",
		gl.Protocol, v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType)

	return &listener{
		Source:            gl,
		Valid:             false,
		Routes:            make(map[types.NamespacedName]*route),
		AcceptedHostnames: make(map[string]struct{}),
		Conditions: []conditions.Condition{
			conditions.NewListenerUnsupportedProtocol(msg),
		},
	}
}

func validateHTTPListener(listener v1beta1.Listener) []conditions.Condition {
	var conds []conditions.Condition

	if listener.Port != 80 {
		msg := fmt.Sprintf("Port %d is not supported for HTTP, use 80", listener.Port)
		conds = append(conds, conditions.NewListenerPortUnavailable(msg))
	}

	// The imported Webhook validation ensures the tls field is not set for an HTTP listener.
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	return conds
}

func validateHTTPSListener(listener v1beta1.Listener, gwNsName string) []conditions.Condition {
	var conds []conditions.Condition

	if listener.Port != 443 {
		msg := fmt.Sprintf("Port %d is not supported for HTTPS, use 443", listener.Port)
		conds = append(conds, conditions.NewListenerPortUnavailable(msg))
	}

	// The imported Webhook validation ensures the tls field is not nil for an HTTPS listener.
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	if *listener.TLS.Mode != v1beta1.TLSModeTerminate {
		msg := fmt.Sprintf("tls.mode %q is not supported, use %q", *listener.TLS.Mode, v1beta1.TLSModeTerminate)
		conds = append(conds, conditions.NewListenerUnsupportedValue(msg))
	}

	if len(listener.TLS.Options) > 0 {
		conds = append(conds, conditions.NewListenerUnsupportedValue("tls.options are not supported"))
	}

	// The imported Webhook validation ensures len(listener.TLS.Certificates) is not 0.
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	certRef := listener.TLS.CertificateRefs[0]

	if certRef.Kind != nil && *certRef.Kind != "Secret" {
		msg := fmt.Sprintf("Kind must be Secret, got %q", *certRef.Kind)
		conds = append(conds, conditions.NewListenerInvalidCertificateRef(msg)...)
	}

	// for Kind Secret, certRef.Group must be nil or empty
	if certRef.Group != nil && *certRef.Group != "" {
		msg := fmt.Sprintf("Group must be empty, got %q", *certRef.Group)
		conds = append(conds, conditions.NewListenerInvalidCertificateRef(msg)...)
	}

	// secret must be in the same namespace as the gateway
	if certRef.Namespace != nil && string(*certRef.Namespace) != gwNsName {
		const msg = "Referenced Secret must belong to the same namespace as the Gateway"
		conds = append(conds, conditions.NewListenerInvalidCertificateRef(msg)...)

	}

	if len(listener.TLS.CertificateRefs) > 1 {
		msg := fmt.Sprintf("Only 1 certificateRef is supported, got %d", len(listener.TLS.CertificateRefs))
		conds = append(conds, conditions.NewListenerUnsupportedValue(msg))
	}

	return conds
}

func validateListenerHostname(host *v1beta1.Hostname) error {
	if host == nil {
		return nil
	}

	h := string(*host)

	if h == "" {
		return nil
	}

	// FIXME(pleshakov): For now, we don't support wildcard hostnames
	if strings.HasPrefix(h, "*") {
		return fmt.Errorf("wildcard hostnames are not supported")
	}

	msgs := validation.IsDNS1123Subdomain(h)
	if len(msgs) > 0 {
		combined := strings.Join(msgs, ",")
		return errors.New(combined)
	}

	return nil
}
