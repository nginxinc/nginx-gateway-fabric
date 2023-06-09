package config

import (
	"strings"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

var mapsTemplate = gotemplate.Must(gotemplate.New("maps").Parse(mapsTemplateText))

func executeMaps(conf dataplane.Configuration) []byte {
	maps := createMaps(append(conf.HTTPServers, conf.SSLServers...))
	return execute(mapsTemplate, maps)
}

func createMaps(servers []dataplane.VirtualServer) []http.Map {
	return buildAddHeaderMaps(servers)
}

func buildAddHeaderMaps(servers []dataplane.VirtualServer) []http.Map {
	addHeaderNames := make(map[string]struct{})

	for _, s := range servers {
		for _, pr := range s.PathRules {
			for _, mr := range pr.MatchRules {
				if mr.Filters.RequestHeaderModifiers != nil {
					for _, addHeader := range mr.Filters.RequestHeaderModifiers.Add {
						lowerName := strings.ToLower(addHeader.Name)
						if _, ok := addHeaderNames[lowerName]; !ok {
							addHeaderNames[lowerName] = struct{}{}
						}
					}
				}
			}
		}
	}

	maps := make([]http.Map, 0, len(addHeaderNames))
	for m := range addHeaderNames {
		maps = append(maps, createAddHeadersMap(m))
	}
	return maps
}

const (
	// In order to prepend any passed client header values to values specified in the add headers field of request
	// header modifiers, we need to create a map parameter regex for any string value
	anyStringFmt = `~.*`
)

func createAddHeadersMap(name string) http.Map {
	underscoreName := convertStringToSafeVariableName(name)
	httpVarSource := "${http_" + underscoreName + "}"
	mapVarName := generateAddHeaderMapVariableName(name)
	params := []http.MapParameter{
		{
			Value:  "default",
			Result: "''",
		},
		{
			Value:  anyStringFmt,
			Result: httpVarSource + ",",
		},
	}
	return http.Map{
		Source:     httpVarSource,
		Variable:   "$" + mapVarName,
		Parameters: params,
	}
}
