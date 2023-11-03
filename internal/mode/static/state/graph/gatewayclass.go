package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

// GatewayClass represents the GatewayClass resource.
type GatewayClass struct {
	// Source is the source resource.
	Source *v1beta1.GatewayClass
	// Conditions include Conditions for the GatewayClass.
	Conditions []conditions.Condition
	// Valid shows whether the GatewayClass is valid.
	Valid bool
}

// processedGatewayClasses holds the resources that belong to NGF.
type processedGatewayClasses struct {
	Winner  *v1beta1.GatewayClass
	Ignored map[types.NamespacedName]*v1beta1.GatewayClass
}

// processGatewayClasses returns the "Winner" GatewayClass, which is defined in
// the command-line argument and references this controller, and a list of "Ignored" GatewayClasses
// that reference this controller, but are not named in the command-line argument.
// Also returns a boolean that says whether or not the GatewayClass defined
// in the command-line argument exists, regardless of which controller it references.
func processGatewayClasses(
	gcs map[types.NamespacedName]*v1beta1.GatewayClass,
	gcName string,
	controllerName string,
) (processedGatewayClasses, bool) {
	processedGwClasses := processedGatewayClasses{}

	var gcExists bool
	for _, gc := range gcs {
		if gc.Name == gcName {
			gcExists = true
			if string(gc.Spec.ControllerName) == controllerName {
				processedGwClasses.Winner = gc
			}
		} else if string(gc.Spec.ControllerName) == controllerName {
			if processedGwClasses.Ignored == nil {
				processedGwClasses.Ignored = make(map[types.NamespacedName]*v1beta1.GatewayClass)
			}
			processedGwClasses.Ignored[client.ObjectKeyFromObject(gc)] = gc
		}
	}

	return processedGwClasses, gcExists
}

func buildGatewayClass(gc *v1beta1.GatewayClass, npCfg *ngfAPI.NginxProxy) *GatewayClass {
	if gc == nil {
		return nil
	}

	var conds []conditions.Condition

	valErr := validateGatewayClass(gc, npCfg)
	if valErr != nil {
		conds = append(conds, staticConds.NewGatewayClassInvalidParameters(valErr.Error()))
	}

	return &GatewayClass{
		Source:     gc,
		Valid:      valErr == nil,
		Conditions: conds,
	}
}

func validateGatewayClass(gc *v1beta1.GatewayClass, npCfg *ngfAPI.NginxProxy) error {
	if gc.Spec.ParametersRef != nil {
		path := field.NewPath("spec").Child("parametersRef")
		if _, ok := supportedParamKinds[string(gc.Spec.ParametersRef.Kind)]; !ok {
			return field.NotSupported(path.Child("kind"), string(gc.Spec.ParametersRef.Kind), []string{"NginxProxy"})
		}

		if gc.Spec.ParametersRef.Namespace == nil {
			return field.Required(path.Child("namespace"), "parametersRef.namespace must be specified for NginxProxy")
		}

		if npCfg == nil {
			return field.NotFound(path.Child("name"), gc.Spec.ParametersRef.Name)
		}
	}

	return nil
}

var supportedParamKinds = map[string]struct{}{
	"NginxProxy": {},
}
