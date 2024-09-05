package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

type SnippetsFilter struct {
	// Source is the SnippetsFilter.
	Source *ngfAPI.SnippetsFilter
	// Conditions define the conditions to be reported in the status of the SnippetsFilter.
	Conditions []conditions.Condition
	// Valid indicates whether the SnippetsFilter is semantically and syntactically valid.
	Valid bool
}

func processSnippetsFilters(
	snippetsFilters map[types.NamespacedName]*ngfAPI.SnippetsFilter,
) map[types.NamespacedName]*SnippetsFilter {
	if len(snippetsFilters) == 0 {
		return nil
	}

	processed := make(map[types.NamespacedName]*SnippetsFilter)

	for nsname, sf := range snippetsFilters {
		processedSf := &SnippetsFilter{
			Source: sf,
			Valid:  true,
		}

		if cond := validateSnippetsFilter(sf); cond != nil {
			processedSf.Valid = false
			processedSf.Conditions = []conditions.Condition{*cond}
		}

		processed[nsname] = processedSf
	}

	return processed
}

func validateSnippetsFilter(filter *ngfAPI.SnippetsFilter) *conditions.Condition {
	var allErrs field.ErrorList
	snippetsPath := field.NewPath("spec.snippets")

	usedContexts := make(map[ngfAPI.NginxContext]struct{})

	for i, snippet := range filter.Spec.Snippets {
		ctxPath := snippetsPath.Index(i).Child("context")

		switch snippet.Context {
		case ngfAPI.NginxContextMain,
			ngfAPI.NginxContextHTTP,
			ngfAPI.NginxContextHTTPServer,
			ngfAPI.NginxContextHTTPServerLocation:
		default:
			err := field.NotSupported(
				ctxPath,
				snippet.Context,
				[]ngfAPI.NginxContext{
					ngfAPI.NginxContextMain,
					ngfAPI.NginxContextHTTP,
					ngfAPI.NginxContextHTTPServer,
					ngfAPI.NginxContextHTTPServerLocation,
				},
			)

			allErrs = append(allErrs, err)
		}

		if _, ok := usedContexts[snippet.Context]; ok {
			allErrs = append(
				allErrs,
				field.Invalid(ctxPath, snippet.Context, "only one snippet is allowed per context"),
			)

			continue
		}

		usedContexts[snippet.Context] = struct{}{}
	}

	if allErrs != nil {
		cond := staticConds.NewSnippetsFilterInvalid(allErrs.ToAggregate().Error())
		return &cond
	}

	return nil
}
