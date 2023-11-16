package graph

import (
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

// Listener represents a Listener of the Gateway resource.
// For now, we only support HTTP and HTTPS listeners.
type Listener struct {
	// Source holds the source of the Listener from the Gateway resource.
	Source v1.Listener
	// Routes holds the routes attached to the Listener.
	// Only valid routes are attached.
	Routes map[types.NamespacedName]*Route
	// AllowedRouteLabelSelector is the label selector for this Listener's allowed routes, if defined.
	AllowedRouteLabelSelector labels.Selector
	// ResolvedSecret is the namespaced name of the Secret resolved for this listener.
	// Only applicable for HTTPS listeners.
	ResolvedSecret *types.NamespacedName
	// Conditions holds the conditions of the Listener.
	Conditions []conditions.Condition
	// SupportedKinds is the list of RouteGroupKinds allowed by the listener.
	SupportedKinds []v1.RouteGroupKind
	// Valid shows whether the Listener is valid.
	// A Listener is considered valid if NGF can generate valid NGINX configuration for it.
	Valid bool
}

func buildListeners(
	gw *v1.Gateway,
	secretResolver *secretResolver,
	refGrantResolver *referenceGrantResolver,
	protectedPorts ProtectedPorts,
) map[string]*Listener {
	listeners := make(map[string]*Listener)

	listenerFactory := newListenerConfiguratorFactory(gw, secretResolver, refGrantResolver, protectedPorts)

	for _, gl := range gw.Spec.Listeners {
		configurator := listenerFactory.getConfiguratorForListener(gl)
		listeners[string(gl.Name)] = configurator.configure(gl)
	}

	return listeners
}

type listenerConfiguratorFactory struct {
	http, https, unsupportedProtocol *listenerConfigurator
}

func (f *listenerConfiguratorFactory) getConfiguratorForListener(l v1.Listener) *listenerConfigurator {
	switch l.Protocol {
	case v1.HTTPProtocolType:
		return f.http
	case v1.HTTPSProtocolType:
		return f.https
	default:
		return f.unsupportedProtocol
	}
}

func newListenerConfiguratorFactory(
	gw *v1.Gateway,
	secretResolver *secretResolver,
	refGrantResolver *referenceGrantResolver,
	protectedPorts ProtectedPorts,
) *listenerConfiguratorFactory {
	sharedPortConflictResolver := createPortConflictResolver()

	return &listenerConfiguratorFactory{
		unsupportedProtocol: &listenerConfigurator{
			validators: []listenerValidator{
				func(listener v1.Listener) []conditions.Condition {
					valErr := field.NotSupported(
						field.NewPath("protocol"),
						listener.Protocol,
						[]string{string(v1.HTTPProtocolType), string(v1.HTTPSProtocolType)},
					)
					return staticConds.NewListenerUnsupportedProtocol(valErr.Error())
				},
			},
		},
		http: &listenerConfigurator{
			validators: []listenerValidator{
				validateListenerAllowedRouteKind,
				validateListenerLabelSelector,
				validateListenerHostname,
				createHTTPListenerValidator(protectedPorts),
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
				createHTTPSListenerValidator(protectedPorts),
			},
			conflictResolvers: []listenerConflictResolver{
				sharedPortConflictResolver,
			},
			externalReferenceResolvers: []listenerExternalReferenceResolver{
				createExternalReferencesForTLSSecretsResolver(gw.Namespace, secretResolver, refGrantResolver),
			},
		},
	}
}

// listenerValidator validates a listener. If the listener is invalid, the validator will return appropriate conditions.
type listenerValidator func(v1.Listener) []conditions.Condition

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

func (c *listenerConfigurator) configure(listener v1.Listener) *Listener {
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
			conds = append(conds, staticConds.NewListenerUnsupportedValue(msg)...)
		}
	}

	supportedKinds := getListenerSupportedKinds(listener)

	if len(conds) > 0 {
		return &Listener{
			Source:         listener,
			Conditions:     conds,
			Valid:          false,
			SupportedKinds: supportedKinds,
		}
	}

	l := &Listener{
		Source:                    listener,
		AllowedRouteLabelSelector: allowedRouteSelector,
		Routes:                    make(map[types.NamespacedName]*Route),
		Valid:                     true,
		SupportedKinds:            supportedKinds,
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

func validateListenerHostname(listener v1.Listener) []conditions.Condition {
	if listener.Hostname == nil {
		return nil
	}

	h := string(*listener.Hostname)

	if h == "" {
		return nil
	}

	if err := validateHostname(h); err != nil {
		path := field.NewPath("hostname")
		valErr := field.Invalid(path, listener.Hostname, err.Error())
		return staticConds.NewListenerUnsupportedValue(valErr.Error())
	}
	return nil
}

func getAndValidateListenerSupportedKinds(listener v1.Listener) (
	[]conditions.Condition,
	[]v1.RouteGroupKind,
) {
	if listener.AllowedRoutes == nil || listener.AllowedRoutes.Kinds == nil {
		return nil, []v1.RouteGroupKind{
			{
				Kind: "HTTPRoute",
			},
		}
	}
	var conds []conditions.Condition

	supportedKinds := make([]v1.RouteGroupKind, 0, len(listener.AllowedRoutes.Kinds))

	validHTTPRouteKind := func(kind v1.RouteGroupKind) bool {
		if kind.Kind != v1.Kind("HTTPRoute") {
			return false
		}
		if kind.Group == nil || *kind.Group != v1.GroupName {
			return false
		}
		return true
	}

	switch listener.Protocol {
	case v1.HTTPProtocolType, v1.HTTPSProtocolType:
		for _, kind := range listener.AllowedRoutes.Kinds {
			if !validHTTPRouteKind(kind) {
				msg := fmt.Sprintf("Unsupported route kind \"%s/%s\"", *kind.Group, kind.Kind)
				conds = append(conds, staticConds.NewListenerInvalidRouteKinds(msg)...)
				continue
			}
			supportedKinds = append(supportedKinds, kind)
		}
	}
	return conds, supportedKinds
}

func validateListenerAllowedRouteKind(listener v1.Listener) []conditions.Condition {
	conds, _ := getAndValidateListenerSupportedKinds(listener)
	return conds
}

func getListenerSupportedKinds(listener v1.Listener) []v1.RouteGroupKind {
	_, kinds := getAndValidateListenerSupportedKinds(listener)
	return kinds
}

func validateListenerLabelSelector(listener v1.Listener) []conditions.Condition {
	if listener.AllowedRoutes != nil &&
		listener.AllowedRoutes.Namespaces != nil &&
		listener.AllowedRoutes.Namespaces.From != nil &&
		*listener.AllowedRoutes.Namespaces.From == v1.NamespacesFromSelector &&
		listener.AllowedRoutes.Namespaces.Selector == nil {
		msg := "Listener's AllowedRoutes Selector must be set when From is set to type Selector"
		return staticConds.NewListenerUnsupportedValue(msg)
	}

	return nil
}

func createHTTPListenerValidator(protectedPorts ProtectedPorts) listenerValidator {
	return func(listener v1.Listener) []conditions.Condition {
		var conds []conditions.Condition

		if err := validateListenerPort(listener.Port, protectedPorts); err != nil {
			path := field.NewPath("port")
			valErr := field.Invalid(path, listener.Port, err.Error())
			conds = append(conds, staticConds.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if listener.TLS != nil {
			panicForBrokenWebhookAssumption(fmt.Errorf("tls is not nil for HTTP listener %q", listener.Name))
		}

		return conds
	}
}

func validateListenerPort(port v1.PortNumber, protectedPorts ProtectedPorts) error {
	if port < 1 || port > 65535 {
		return errors.New("port must be between 1-65535")
	}

	if portName, ok := protectedPorts[int32(port)]; ok {
		return fmt.Errorf("port is already in use as %v", portName)
	}

	return nil
}

func createHTTPSListenerValidator(protectedPorts ProtectedPorts) listenerValidator {
	return func(listener v1.Listener) []conditions.Condition {
		var conds []conditions.Condition

		if err := validateListenerPort(listener.Port, protectedPorts); err != nil {
			path := field.NewPath("port")
			valErr := field.Invalid(path, listener.Port, err.Error())
			conds = append(conds, staticConds.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if listener.TLS == nil {
			panicForBrokenWebhookAssumption(fmt.Errorf("tls is nil for HTTPS listener %q", listener.Name))
		}

		tlsPath := field.NewPath("tls")

		if *listener.TLS.Mode != v1.TLSModeTerminate {
			valErr := field.NotSupported(
				tlsPath.Child("mode"),
				*listener.TLS.Mode,
				[]string{string(v1.TLSModeTerminate)},
			)
			conds = append(conds, staticConds.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if len(listener.TLS.Options) > 0 {
			path := tlsPath.Child("options")
			valErr := field.Forbidden(path, "options are not supported")
			conds = append(conds, staticConds.NewListenerUnsupportedValue(valErr.Error())...)
		}

		if len(listener.TLS.CertificateRefs) == 0 {
			panicForBrokenWebhookAssumption(fmt.Errorf("zero certificateRefs for HTTPS listener %q", listener.Name))
		}

		certRef := listener.TLS.CertificateRefs[0]

		certRefPath := tlsPath.Child("certificateRefs").Index(0)

		if certRef.Kind != nil && *certRef.Kind != "Secret" {
			path := certRefPath.Child("kind")
			valErr := field.NotSupported(path, *certRef.Kind, []string{"Secret"})
			conds = append(conds, staticConds.NewListenerInvalidCertificateRef(valErr.Error())...)
		}

		// for Kind Secret, certRef.Group must be nil or empty
		if certRef.Group != nil && *certRef.Group != "" {
			path := certRefPath.Child("group")
			valErr := field.NotSupported(path, *certRef.Group, []string{""})
			conds = append(conds, staticConds.NewListenerInvalidCertificateRef(valErr.Error())...)
		}

		if l := len(listener.TLS.CertificateRefs); l > 1 {
			path := tlsPath.Child("certificateRefs")
			valErr := field.TooMany(path, l, 1)
			conds = append(conds, staticConds.NewListenerUnsupportedValue(valErr.Error())...)
		}

		return conds
	}
}

func createPortConflictResolver() listenerConflictResolver {
	conflictedPorts := make(map[v1.PortNumber]bool)
	portProtocolOwner := make(map[v1.PortNumber]v1.ProtocolType)
	listenersByPort := make(map[v1.PortNumber][]*Listener)

	format := "Multiple listeners for the same port %d specify incompatible protocols; " +
		"ensure only one protocol per port"

	return func(l *Listener) {
		port := l.Source.Port

		// if port is in map of conflictedPorts then we only need to set the current listener to invalid
		if conflictedPorts[port] {
			l.Valid = false

			conflictedConds := staticConds.NewListenerProtocolConflict(fmt.Sprintf(format, port))
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
				conflictedConds := staticConds.NewListenerProtocolConflict(fmt.Sprintf(format, port))
				l.Conditions = append(l.Conditions, conflictedConds...)
			}
		}
	}
}

func createExternalReferencesForTLSSecretsResolver(
	gwNs string,
	secretResolver *secretResolver,
	refGrantResolver *referenceGrantResolver,
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
			if !refGrantResolver.refAllowed(toSecret(certRefNsName), fromGateway(gwNs)) {
				msg := fmt.Sprintf("Certificate ref to secret %s not permitted by any ReferenceGrant", certRefNsName)

				l.Conditions = append(l.Conditions, staticConds.NewListenerRefNotPermitted(msg)...)
				l.Valid = false
				return
			}
		}

		if err := secretResolver.resolve(certRefNsName); err != nil {
			path := field.NewPath("tls", "certificateRefs").Index(0)
			// field.NotFound could be better, but it doesn't allow us to set the error message.
			valErr := field.Invalid(path, certRefNsName, err.Error())

			l.Conditions = append(l.Conditions, staticConds.NewListenerInvalidCertificateRef(valErr.Error())...)
			l.Valid = false
		} else {
			l.ResolvedSecret = &certRefNsName
		}
	}
}

// GetAllowedRouteLabelSelector returns a listener's AllowedRoutes label selector if it exists.
func GetAllowedRouteLabelSelector(l v1.Listener) *metav1.LabelSelector {
	if l.AllowedRoutes != nil && l.AllowedRoutes.Namespaces != nil {
		if *l.AllowedRoutes.Namespaces.From == v1.NamespacesFromSelector &&
			l.AllowedRoutes.Namespaces.Selector != nil {
			return l.AllowedRoutes.Namespaces.Selector
		}
	}

	return nil
}
