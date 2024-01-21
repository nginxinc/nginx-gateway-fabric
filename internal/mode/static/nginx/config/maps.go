package config

import (
	"strings"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var mapsTemplate = gotemplate.Must(gotemplate.New("maps").Parse(mapsTemplateText))

func executeMaps(conf dataplane.Configuration) []executeResult {
	maps := buildAddHeaderMaps(append(conf.HTTPServers, conf.SSLServers...))
	result := executeResult{
		dest: httpConfigFile,
		data: execute(mapsTemplate, maps),
	}
	return []executeResult{result}
}

func buildAddHeaderMaps(servers []dataplane.VirtualServer) []http.Map {
	addHeaderNames := make(map[string]struct{})
	extractAddHeaderNames := func(headerModifiers []dataplane.HTTPHeader) {
		for _, addHeader := range headerModifiers {
			lowerName := strings.ToLower(addHeader.Name)
			if _, ok := addHeaderNames[lowerName]; !ok {
				addHeaderNames[lowerName] = struct{}{}
			}
		}
	}

	for _, s := range servers {
		for _, pr := range s.PathRules {
			for _, mr := range pr.MatchRules {
				if mr.Filters.RequestHeaderModifiers != nil {
					extractAddHeaderNames(mr.Filters.RequestHeaderModifiers.Add)
				}
				if mr.Filters.ResponseHeaderModifiers != nil {
					extractAddHeaderNames(mr.Filters.ResponseHeaderModifiers.Add)
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
