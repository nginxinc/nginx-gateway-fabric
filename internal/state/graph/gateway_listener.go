package graph

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

// Listener represents a Listener of the Gateway resource.
// For now, we only support HTTP and HTTPS listeners.
type Listener struct {
	// Source holds the source of the Listener from the Gateway resource.
	Source v1beta1.Listener
	// Routes holds the routes attached to the Listener.
	// Only valid routes are attached.
	Routes map[types.NamespacedName]*Route
	// AllowedRouteLabelSelector is the label selector for this Listener's allowed routes, if defined.
	AllowedRouteLabelSelector labels.Selector
	// SecretPath is the path to the secret on disk.
	SecretPath string
	// Conditions holds the conditions of the Listener.
	Conditions []conditions.Condition
	// Valid shows whether the Listener is valid.
	// A Listener is considered valid if NKG can generate valid NGINX configuration for it.
	Valid bool
}

func buildListeners(
	gw *v1beta1.Gateway,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
	refGrants map[types.NamespacedName]*v1beta1.ReferenceGrant,
) map[string]*Listener {
	listeners := make(map[string]*Listener)

	listenerFactory := newListenerConfiguratorFactory(gw, secretMemoryMgr, refGrants)

	for _, gl := range gw.Spec.Listeners {
		configurator := listenerFactory.getConfiguratorForListener(gl)
		listeners[string(gl.Name)] = configurator.configure(gl)
	}

	return listeners
}

type listenerConfiguratorFactory struct {
	http, https, unsupportedProtocol *listenerConfigurator
}

func (f *listenerConfiguratorFactory) getConfiguratorForListener(l v1beta1.Listener) *listenerConfigurator {
	switch l.Protocol {
	case v1beta1.HTTPProtocolType:
		return f.http
	case v1beta1.HTTPSProtocolType:
		return f.https
	default:
		return f.unsupportedProtocol
	}
}

func newListenerConfiguratorFactory(
	gw *v1beta1.Gateway,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
	refGrants map[types.NamespacedName]*v1beta1.ReferenceGrant,
) *listenerConfiguratorFactory {
	sharedPortConflictResolver := createPortConflictResolver()

	return &listenerConfiguratorFactory{
		unsupportedProtocol: &listenerConfigurator{
			validators: []listenerValidator{
				func(listener v1beta1.Listener) []conditions.Condition {
					valErr := field.NotSupported(
						field.NewPath("protocol"),
						listener.Protocol,
						[]string{string(v1beta1.HTTPProtocolType), string(v1beta1.HTTPSProtocolType)},
					)
					return conditions.NewListenerUnsupportedProtocol(valErr.Error())
				},
			},
		},
		http: &listenerConfigurator{
			validators: []listenerValidator{
				validateListenerAllowedRouteKind,
				validateListenerLabelSelector,
				validateListenerHostname,
				validateHTTPListener,
			},
			conflictResolvers: []listenerConflictResolver{
				sharedPortConflictResolver,
			},
		},
		https: &listenerConfigurator{
			validators: []listenerValidator{
				validateListenerAllowedRouteKind,
				validateListenerLabelSelector,
				validateListenerHostname,
				createHTTPSListenerValidator(),
			},
			conflictResolvers: []listenerConflictResolver{
				sharedPortConflictResolver,
			},
			externalReferenceResolvers: []listenerExternalReferenceResolver{
				createExternalReferencesForTLSSecretsResolver(gw.Namespace, secretMemoryMgr, refGrants),
			},
		},
	}
}

// listenerValidator validates a listener. If the listener is invalid, the validator will return appropriate conditions.
type listenerValidator func(v1beta1.Listener) []conditions.Condition

// listenerConflictResolver resolves conflicts between listeners. In case of a conflict, the resolver will make
// the conflicting listeners invalid - i.e. it will modify the passed listener and the previously processed conflicting
// listener. It will also add appropriate conditions to the listeners.
type listenerConflictResolver func(listener *Listener)

// listenerExternalReferenceResolver resolves external references for a listener. If the reference is not resolvable,
// the resolver will make the listener invalid and add appropriate conditions.
type listenerExternalReferenceResolver func(listener *Listener)

// listenerConfigurator is responsible for configuring a listener.
// validators, conflictResolvers, externalReferenceResolvers generate conditions for invalid fields of the listener.
// Because the Gateway status includes a status field for each listener, the messages in those conditions
// don't need to include the full path to the field (e.g. "spec.listeners[0].hostname"). They will include
// a path starting from the field of a listener (e.g. "hostname", "tls.options").
type listenerConfigurator struct {
	// validators must not depend on the order of execution.
	validators []listenerValidator

	// conflictResolvers can depend on validators - they will only be executed if all validators pass.
	conflictResolvers []listenerConflictResolver
	// externalReferenceResolvers can depend on validators - they will only be executed if all validators pass.
	externalReferenceResolvers []listenerExternalReferenceResolver
}

func (c *listenerConfigurator) configure(listener v1beta1.Listener) *Listener {
	var conds []conditions.Condition

	// validators might return different conditions, so we run them all.
	for _, validator := range c.validators {
		conds = append(conds, validator(listener)...)
	}

	var allowedRouteSelector labels.Selector
	if selector := GetAllowedRouteLabelSelector(listener); selector != nil {
		var err error
		allowedRouteSelector, err = metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			msg := fmt.Sprintf("invalid label selector: %s", err.Error())
			conds = append(conds, conditions.NewListenerUnsupportedValue(msg)...)
		}
	}

	if len(conds) > 0 {
		return &Listener{
			Source:     listener,
			Conditions: conds,
			Valid:      false,
		}
	}

	l := &Listener{
		Source:                    listener,
		AllowedRouteLabelSelector: allowedRouteSelector,
		Routes:                    make(map[types.NamespacedName]*Route),
		Valid:                     true,
	}

	// resolvers might add different conditions to the listener, so we run them all.

	for _, resolver := range c.conflictResolvers {
		resolver(l)
	}

	for _, resolver := range c.externalReferenceResolvers {
		resolver(l)
	}

	return l
}

func validateListenerHostname(listener v1beta1.Listener) []conditions.Condition {
	if listener.Hostname == nil {
		return nil
	}

	h := string(*listener.Hostname)

	if h == "" {
		return nil
	}

	err := validateHostname(h)
	if err != nil {
		path := field.NewPath("hostname")
		valErr := field.Invalid(path, listener.Hostname, err.Error())
		return conditions.NewListenerUnsupportedValue(valErr.Error())
	}
	return nil
}

func validateListenerAllowedRouteKind(listener v1beta1.Listener) []conditions.Condition {
	validHTTPRouteKind := func(kind v1beta1.RouteGroupKind) bool {
		if kind.Kind != v1beta1.Kind("HTTPRoute") {
			return false
		}
		if kind.Group == nil || *kind.Group != v1beta1.GroupName {
			return false
		}
		return true
	}

	switch listener.Protocol {
	case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
		if listener.AllowedRoutes != nil {
			for _, kind := range listener.AllowedRoutes.Kinds {
				if !validHTTPRouteKind(kind) {
					msg := fmt.Sprintf("Unsupported route kind \"%s/%s\"", *kind.Group, kind.Kind)
					return conditions.NewListenerUnsupportedValue(msg)
				}
			}
		}
	}

	return nil
}

func validateListenerLabelSelector(listener v1beta1.Listener) []conditions.Condition {
	if listener.AllowedRoutes != nil &&
		listener.AllowedRoutes.Namespaces != nil &&
		listener.AllowedRoutes.Namespaces.From != nil &&
		*listener.AllowedRoutes.Namespaces.From == v1beta1.NamespacesFromSelector &&
		listener.AllowedRoutes.Namespaces.Selector == nil {
		msg := "Listener's AllowedRoutes Selector must be set when From is set to type Selector"
		return conditions.NewListenerUnsupportedValue(msg)
	}

	return nil
}

func validateHTTPListener(listener v1beta1.Listener) []conditions.Condition {
	if err := validateListenerPort(listener.Port); err != nil {
		path := field.NewPath("port")
		valErr := field.Invalid(path, listener.Port, err.Error())
		return conditions.NewListenerUnsupportedValue(valErr.Error())
	}

	if listener.TLS != nil {
		panicForBrokenWebhookAssumption(fmt.Errorf("tls is not nil for HTTP listener %q", listener.Name))
	}

	return nil
}

func validateListenerPort(port v1beta1.PortNumber) error {
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1-65535")
	}

	return nil
}

func createHTTPSListenerValidator() listenerValidator {
	return func(listener v1beta1.Listener) []conditions.Condition {
		var conds []conditions.Condition

		if err := validateListenerPort(listener.Port); err != nil {
			path := field.NewPath("port")
			valErr := field.Invalid(path, listener.Port, err.Error())
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if listener.TLS == nil {
			panicForBrokenWebhookAssumption(fmt.Errorf("tls is nil for HTTPS listener %q", listener.Name))
		}

		tlsPath := field.NewPath("tls")

		if *listener.TLS.Mode != v1beta1.TLSModeTerminate {
			valErr := field.NotSupported(
				tlsPath.Child("mode"),
				*listener.TLS.Mode,
				[]string{string(v1beta1.TLSModeTerminate)},
			)
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if len(listener.TLS.Options) > 0 {
			path := tlsPath.Child("options")
			valErr := field.Forbidden(path, "options are not supported")
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if len(listener.TLS.CertificateRefs) == 0 {
			panicForBrokenWebhookAssumption(fmt.Errorf("zero certificateRefs for HTTPS listener %q", listener.Name))
		}

		certRef := listener.TLS.CertificateRefs[0]

		certRefPath := tlsPath.Child("certificateRefs").Index(0)

		if certRef.Kind != nil && *certRef.Kind != "Secret" {
			path := certRefPath.Child("kind")
			valErr := field.NotSupported(path, *certRef.Kind, []string{"Secret"})
			conds = append(conds, conditions.NewListenerInvalidCertificateRef(valErr.Error())...)
		}

		// for Kind Secret, certRef.Group must be nil or empty
		if certRef.Group != nil && *certRef.Group != "" {
			path := certRefPath.Child("group")
			valErr := field.NotSupported(path, *certRef.Group, []string{""})
			conds = append(conds, conditions.NewListenerInvalidCertificateRef(valErr.Error())...)
		}

		if l := len(listener.TLS.CertificateRefs); l > 1 {
			path := tlsPath.Child("certificateRefs")
			valErr := field.TooMany(path, l, 1)
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error())...)
		}

		return conds
	}
}

func createPortConflictResolver() listenerConflictResolver {
	conflictedPorts := make(map[v1beta1.PortNumber]bool)
	portProtocolOwner := make(map[v1beta1.PortNumber]v1beta1.ProtocolType)
	listenersByPort := make(map[v1beta1.PortNumber][]*Listener)

	format := "Multiple listeners for the same port %d specify incompatible protocols; " +
		"ensure only one protocol per port"

	return func(l *Listener) {
		port := l.Source.Port

		// if port is in map of conflictedPorts then we only need to set the current listener to invalid
		if conflictedPorts[port] {
			l.Valid = false

			conflictedConds := conditions.NewListenerProtocolConflict(fmt.Sprintf(format, port))
			l.Conditions = append(l.Conditions, conflictedConds...)
			return
		}

		// otherwise, we add the listener to the list of listeners for this port
		// and then check if the protocol owner for the port is different from the current listener's protocol.

		listenersByPort[port] = append(listenersByPort[port], l)

		protocol, ok := portProtocolOwner[port]
		if !ok {
			portProtocolOwner[port] = l.Source.Protocol
			return
		}

		// if protocol owner doesn't match the listener's protocol we mark the port as conflicted,
		// and invalidate all listeners we've seen for this port.
		if protocol != l.Source.Protocol {
			conflictedPorts[port] = true
			for _, l := range listenersByPort[port] {
				l.Valid = false
				conflictedConds := conditions.NewListenerProtocolConflict(fmt.Sprintf(format, port))
				l.Conditions = append(l.Conditions, conflictedConds...)
			}
		}
	}
}

func createExternalReferencesForTLSSecretsResolver(
	gwNs string,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
	refGrants map[types.NamespacedName]*v1beta1.ReferenceGrant,
) listenerExternalReferenceResolver {
	return func(l *Listener) {
		certRef := l.Source.TLS.CertificateRefs[0]

		certRefNs := gwNs
		if certRef.Namespace != nil {
			certRefNs = string(*certRef.Namespace)
		}

		certRefNsName := types.NamespacedName{
			Namespace: certRefNs,
			Name:      string(certRef.Name),
		}

		if certRefNs != gwNs {
			if !refGrantAllowsGatewayToSecret(refGrants, gwNs, certRefNsName) {
				msg := fmt.Sprintf(
					"Certificate ref to secret %s/%s not permitted by any ReferenceGrant",
					certRefNs,
					certRef.Name,
				)

				l.Conditions = append(l.Conditions, conditions.NewListenerRefNotPermitted(msg)...)
				l.Valid = false
				return
			}
		}

		var err error

		l.SecretPath, err = secretMemoryMgr.Request(certRefNsName)
		if err != nil {
			path := field.NewPath("tls", "certificateRefs").Index(0)
			// field.NotFound could be better, but it doesn't allow us to set the error message.
			valErr := field.Invalid(path, certRefNsName, err.Error())

			l.Conditions = append(l.Conditions, conditions.NewListenerInvalidCertificateRef(valErr.Error())...)
			l.Valid = false
		}
	}
}

// GetAllowedRouteLabelSelector returns a listener's AllowedRoutes label selector if it exists.
func GetAllowedRouteLabelSelector(l v1beta1.Listener) *metav1.LabelSelector {
	if l.AllowedRoutes != nil && l.AllowedRoutes.Namespaces != nil {
		if *l.AllowedRoutes.Namespaces.From == v1beta1.NamespacesFromSelector &&
			l.AllowedRoutes.Namespaces.Selector != nil {
			return l.AllowedRoutes.Namespaces.Selector
		}
	}

	return nil
}
