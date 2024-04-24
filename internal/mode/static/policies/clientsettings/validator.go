package clientsettings

import (
	"fmt"
	"slices"

	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
)

// Validator validates a ClientSettingsPolicy.
// Implements policies.Validator interface.
type Validator struct{}

// Validate validates the spec of a ClientSettingsPolicy.
func (c Validator) Validate(policy policies.Policy) error {
	csp, ok := policy.(*ngfAPI.ClientSettingsPolicy)
	if !ok {
		panic(fmt.Sprintf("expected ClientSettingsPolicy, got: %T", policy))
	}

	if err := validateTargetRef(csp.Spec.TargetRef, csp.Namespace); err != nil {
		return err
	}

	return validateSettings(csp.Spec)
}

// Conflicts returns true if the two ClientSettingsPolicies conflict.
func (c Validator) Conflicts(polA, polB policies.Policy) bool {
	a, okA := polA.(*ngfAPI.ClientSettingsPolicy)
	b, okB := polB.(*ngfAPI.ClientSettingsPolicy)

	if !okA || !okB {
		panic(fmt.Sprintf("expected ClientSettingsPolicy, got: %T", polA))
	}

	return conflicts(a.Spec, b.Spec)
}

func conflicts(a, b ngfAPI.ClientSettingsPolicySpec) bool {
	if a.Body != nil && b.Body != nil {
		if a.Body.Timeout != nil && b.Body.Timeout != nil {
			return true
		}

		if a.Body.MaxSize != nil && b.Body.MaxSize != nil {
			return true
		}
	}

	if a.KeepAlive != nil && b.KeepAlive != nil {
		if a.KeepAlive.Requests != nil && b.KeepAlive.Requests != nil {
			return true
		}

		if a.KeepAlive.Time != nil && b.KeepAlive.Time != nil {
			return true
		}

		if a.KeepAlive.Timeout != nil && b.KeepAlive.Timeout != nil {
			return true
		}
	}

	return false
}

func validateTargetRef(ref v1alpha2.PolicyTargetReference, policyNs string) error {
	basePath := field.NewPath("spec").Child("targetRef")

	if ref.Namespace != nil && string(*ref.Namespace) != policyNs {
		path := basePath.Child("namespace")

		return field.Invalid(path, *ref.Namespace, "targetRef must be in the same namespace as the policy")
	}

	if ref.Group != gatewayv1.GroupName {
		path := basePath.Child("group")

		return field.Invalid(
			path,
			ref.Group,
			fmt.Sprintf("unsupported targetRef Group %q; must be %s", ref.Group, gatewayv1.GroupName),
		)
	}

	kinds := []gatewayv1.Kind{"HTTPRoute", "Gateway"}

	if !slices.Contains(kinds, ref.Kind) {
		path := basePath.Child("kind")

		return field.Invalid(
			path,
			ref.Kind,
			fmt.Sprintf("unsupported targetRef Kind %q; Kind must be one of: %v", ref.Kind, kinds),
		)
	}

	return nil
}

func validateSettings(spec ngfAPI.ClientSettingsPolicySpec) error {
	var allErrs field.ErrorList
	fieldPath := field.NewPath("spec")

	if spec.Body != nil {
		if err := policies.ValidateDuration(spec.Body.Timeout); err != nil {
			path := fieldPath.Child("body").Child("timeout")

			allErrs = append(allErrs, field.Invalid(path, *spec.Body.Timeout, err.Error()))
		}

		if err := policies.ValidateSize(spec.Body.MaxSize); err != nil {
			path := fieldPath.Child("body").Child("size")

			allErrs = append(allErrs, field.Invalid(path, *spec.Body.MaxSize, err.Error()))
		}
	}

	if spec.KeepAlive != nil {
		if spec.KeepAlive.Requests != nil {
			requests := *spec.KeepAlive.Requests
			if requests < 0 {
				path := fieldPath.Child("keepAlive").Child("requests")

				allErrs = append(
					allErrs,
					field.Invalid(path, *spec.KeepAlive.Requests, "requests is invalid: must be positive"),
				)
			}
		}

		if err := policies.ValidateDuration(spec.KeepAlive.Time); err != nil {
			path := fieldPath.Child("body").Child("keepAlive").Child("time")

			allErrs = append(allErrs, field.Invalid(path, *spec.KeepAlive.Time, err.Error()))
		}

		if spec.KeepAlive.Timeout != nil {
			timeout := spec.KeepAlive.Timeout

			if err := policies.ValidateDuration(timeout.Server); err != nil {
				path := fieldPath.Child("keepAlive").Child("timeout").Child("server")

				allErrs = append(
					allErrs,
					field.Invalid(path, *spec.KeepAlive.Timeout.Server, err.Error()),
				)
			}

			if err := policies.ValidateDuration(timeout.Header); err != nil {
				path := fieldPath.Child("keepAlive").Child("timeout").Child("header")

				allErrs = append(
					allErrs,
					field.Invalid(path, *spec.KeepAlive.Timeout.Header, err.Error()),
				)
			}

			if spec.KeepAlive.Timeout.Header != nil && spec.KeepAlive.Timeout.Server == nil {
				path := fieldPath.Child("keepAlive").Child("timeout")

				allErrs = append(
					allErrs,
					field.Invalid(
						path,
						nil,
						"server timeout must be set if header timeout is set",
					),
				)
			}
		}
	}
	return allErrs.ToAggregate()
}
