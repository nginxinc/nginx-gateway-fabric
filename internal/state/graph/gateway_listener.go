package graph

import (
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
) map[string]*Listener {
	listeners := make(map[string]*Listener)

	listenerFactory := newListenerConfiguratorFactory(gw, secretMemoryMgr)

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
) *listenerConfiguratorFactory {
	return &listenerConfiguratorFactory{
		unsupportedProtocol: &listenerConfigurator{
			validators: []listenerValidator{
				func(listener v1beta1.Listener) []conditions.Condition {
					valErr := field.NotSupported(
						field.NewPath("protocol"),
						listener.Protocol,
						[]string{string(v1beta1.HTTPProtocolType), string(v1beta1.HTTPSProtocolType)},
					)
					return []conditions.Condition{conditions.NewListenerUnsupportedProtocol(valErr.Error())}
				},
			},
		},
		http: &listenerConfigurator{
			validators: []listenerValidator{
				validateListenerAllowedRouteKind,
				validateListenerHostname,
				validateHTTPListener,
			},
			conflictResolvers: []listenerConflictResolver{
				createHostnameConflictResolver(),
			},
		},
		https: &listenerConfigurator{
			validators: []listenerValidator{
				validateListenerAllowedRouteKind,
				validateListenerHostname,
				createHTTPSListenerValidator(gw.Namespace),
			},
			conflictResolvers: []listenerConflictResolver{
				createHostnameConflictResolver(),
			},
			externalReferenceResolvers: []listenerExternalReferenceResolver{
				createExternalReferencesForTLSSecretsResolver(gw.Namespace, secretMemoryMgr),
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
			conds = append(conds, conditions.NewListenerUnsupportedValue(msg))
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
		return []conditions.Condition{conditions.NewListenerUnsupportedValue(valErr.Error())}
	}
	return nil
}

func validateListenerAllowedRouteKind(listener v1beta1.Listener) []conditions.Condition {
	switch listener.Protocol {
	case v1beta1.HTTPProtocolType, v1beta1.HTTPSProtocolType:
		if listener.AllowedRoutes != nil {
			for _, kind := range listener.AllowedRoutes.Kinds {
				if kind.Kind != v1beta1.Kind("HTTPRoute") || *kind.Group != v1beta1.GroupName {
					return []conditions.Condition{
						conditions.NewListenerUnsupportedValue(fmt.Sprintf("Unsupported route kind \"%s/%s\"", *kind.Group, kind.Kind)),
					}
				}
			}
		}
	}

	return nil
}

func validateHTTPListener(listener v1beta1.Listener) []conditions.Condition {
	if listener.Port != 80 {
		path := field.NewPath("port")
		valErr := field.NotSupported(path, listener.Port, []string{"80"})
		return []conditions.Condition{conditions.NewListenerPortUnavailable(valErr.Error())}
	}

	if listener.TLS != nil {
		panicForBrokenWebhookAssumption(fmt.Errorf("tls is not nil for HTTP listener %q", listener.Name))
	}

	return nil
}

func createHTTPSListenerValidator(gwNsName string) listenerValidator {
	return func(listener v1beta1.Listener) []conditions.Condition {
		var conds []conditions.Condition

		if listener.Port != 443 {
			path := field.NewPath("port")
			valErr := field.NotSupported(path, listener.Port, []string{"443"})
			conds = append(conds, conditions.NewListenerPortUnavailable(valErr.Error()))
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
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error()))
		}

		if len(listener.TLS.Options) > 0 {
			path := tlsPath.Child("options")
			valErr := field.Forbidden(path, "options are not supported")
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error()))
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

		// secret must be in the same namespace as the gateway
		if certRef.Namespace != nil && string(*certRef.Namespace) != gwNsName {
			const detail = "Referenced Secret must belong to the same namespace as the Gateway"
			path := certRefPath.Child("namespace")
			valErr := field.Invalid(path, certRef.Namespace, detail)
			conds = append(conds, conditions.NewListenerInvalidCertificateRef(valErr.Error())...)
		}

		if l := len(listener.TLS.CertificateRefs); l > 1 {
			path := tlsPath.Child("certificateRefs")
			valErr := field.TooMany(path, l, 1)
			conds = append(conds, conditions.NewListenerUnsupportedValue(valErr.Error()))
		}

		return conds
	}
}

func createHostnameConflictResolver() listenerConflictResolver {
	usedHostnames := make(map[string]*Listener)

	return func(l *Listener) {
		h := getHostname(l.Source.Hostname)

		if holder, exist := usedHostnames[h]; exist {
			l.Valid = false

			holder.Valid = false // all listeners for the same hostname become conflicted

			format := "Multiple listeners for the same port use the same hostname %q; " +
				"ensure only one listener uses that hostname"
			conflictedConds := conditions.NewListenerConflictedHostname(fmt.Sprintf(format, h))

			holder.Conditions = append(holder.Conditions, conflictedConds...)
			l.Conditions = append(l.Conditions, conflictedConds...)

			return
		}

		usedHostnames[h] = l
	}
}

func createExternalReferencesForTLSSecretsResolver(
	gwNs string,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
) listenerExternalReferenceResolver {
	return func(l *Listener) {
		nsname := types.NamespacedName{
			Namespace: gwNs,
			Name:      string(l.Source.TLS.CertificateRefs[0].Name),
		}

		var err error

		l.SecretPath, err = secretMemoryMgr.Request(nsname)
		if err != nil {
			path := field.NewPath("tls", "certificateRefs").Index(0)
			// field.NotFound could be better, but it doesn't allow us to set the error message.
			valErr := field.Invalid(path, nsname, err.Error())

			l.Conditions = append(l.Conditions, conditions.NewListenerInvalidCertificateRef(valErr.Error())...)
			l.Valid = false
		}
	}
}

// GetAllowedRouteLabelSelector returns a listener's AllowedRoutes label selector if it exists.
func GetAllowedRouteLabelSelector(l v1beta1.Listener) *metav1.LabelSelector {
	if l.AllowedRoutes != nil && l.AllowedRoutes.Namespaces != nil && l.AllowedRoutes.Namespaces.Selector != nil {
		return l.AllowedRoutes.Namespaces.Selector
	}

	return nil
}
