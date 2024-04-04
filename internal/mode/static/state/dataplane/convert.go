package dataplane

import (
	"fmt"

	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func convertMatch(m v1.HTTPRouteMatch) Match {
	match := Match{}

	if m.Method != nil {
		method := string(*m.Method)
		match.Method = &method
	}

	if len(m.Headers) != 0 {
		match.Headers = make([]HTTPHeaderMatch, 0, len(m.Headers))
		for _, h := range m.Headers {
			match.Headers = append(match.Headers, HTTPHeaderMatch{
				Name:  string(h.Name),
				Value: h.Value,
			})
		}
	}

	if len(m.QueryParams) != 0 {
		match.QueryParams = make([]HTTPQueryParamMatch, 0, len(m.QueryParams))
		for _, q := range m.QueryParams {
			match.QueryParams = append(match.QueryParams, HTTPQueryParamMatch{
				Name:  string(q.Name),
				Value: q.Value,
			})
		}
	}

	return match
}

func convertHTTPRequestRedirectFilter(filter *v1.HTTPRequestRedirectFilter) *HTTPRequestRedirectFilter {
	return &HTTPRequestRedirectFilter{
		Scheme:     filter.Scheme,
		Hostname:   (*string)(filter.Hostname),
		Port:       (*int32)(filter.Port),
		StatusCode: filter.StatusCode,
	}
}

func convertHTTPURLRewriteFilter(filter *v1.HTTPURLRewriteFilter) *HTTPURLRewriteFilter {
	return &HTTPURLRewriteFilter{
		Hostname: (*string)(filter.Hostname),
		Path:     convertPathModifier(filter.Path),
	}
}

func convertHTTPHeaderFilter(filter *v1.HTTPHeaderFilter) *HTTPHeaderFilter {
	result := &HTTPHeaderFilter{
		Remove: filter.Remove,
	}

	if len(filter.Set) != 0 {
		result.Set = make([]HTTPHeader, 0, len(filter.Set))
		for _, s := range filter.Set {
			result.Set = append(result.Set, HTTPHeader{Name: string(s.Name), Value: s.Value})
		}
	}

	if len(filter.Add) != 0 {
		result.Add = make([]HTTPHeader, 0, len(filter.Add))
		for _, a := range filter.Add {
			result.Add = append(result.Add, HTTPHeader{Name: string(a.Name), Value: a.Value})
		}
	}

	return result
}

func convertHTTPRequestMirrorFilter(filter *v1.HTTPRequestMirrorFilter) *HTTPRequestMirrorFilter {
	result := &HTTPRequestMirrorFilter{}

	filterBackendRef := filter.BackendRef

	backendRefName := (string)(filterBackendRef.Name)
	if len(backendRefName) > 0 {
		result.Name = helpers.GetPointer(backendRefName)
	}

	backendRefNamespace := (*string)(filterBackendRef.Namespace)
	if backendRefNamespace != nil && len(*backendRefNamespace) > 0 {
		result.Namespace = backendRefNamespace
	}

	port := (*int32)(filterBackendRef.Port)
	if port != nil && *port > 0 {
		result.Port = port
	}

	// this initialization is partial and must be completed with the path
	return result
}

func updateHTTPMirrorFilterRoute(path *v1.HTTPPathMatch, f *HTTPRequestMirrorFilter) {
	if f.Name != nil {
		f.Target = helpers.CreateMirrorBackendPath(path.Value, f.Namespace, f.Name)
	}
}

func convertPathType(pathType v1.PathMatchType) PathType {
	switch pathType {
	case v1.PathMatchPathPrefix:
		return PathTypePrefix
	case v1.PathMatchExact:
		return PathTypeExact
	default:
		panic(fmt.Sprintf("unsupported path type: %s", pathType))
	}
}

func convertPathModifier(path *v1.HTTPPathModifier) *HTTPPathModifier {
	if path != nil {
		switch path.Type {
		case v1.FullPathHTTPPathModifier:
			return &HTTPPathModifier{
				Type:        ReplaceFullPath,
				Replacement: *path.ReplaceFullPath,
			}
		case v1.PrefixMatchHTTPPathModifier:
			return &HTTPPathModifier{
				Type:        ReplacePrefixMatch,
				Replacement: *path.ReplacePrefixMatch,
			}
		}
	}

	return nil
}
