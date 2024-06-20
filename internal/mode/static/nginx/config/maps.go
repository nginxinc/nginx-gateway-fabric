package config

import (
	"fmt"
	"strings"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var mapsTemplate = gotemplate.Must(gotemplate.New("maps").Parse(mapsTemplateText))

func executeMaps(conf dataplane.Configuration) []executeResult {
	maps := buildAddHeaderMaps(append(conf.HTTPServers, conf.SSLServers...))
	result := executeResult{
		dest: httpConfigFile,
		data: helpers.MustExecuteTemplate(mapsTemplate, maps),
	}

	return []executeResult{result}
}

func executeStreamMaps(conf dataplane.Configuration) []executeResult {
	maps := createStreamMaps(conf)

	result := executeResult{
		dest: streamConfigFile,
		data: helpers.MustExecuteTemplate(mapsTemplate, maps),
	}

	return []executeResult{result}
}

func createStreamMaps(conf dataplane.Configuration) []*http.Map {
	var maps []*http.Map
	portsToMap := make(map[int32]*http.Map)

	for _, t := range conf.TLSServers {
		streamMap, ok := portsToMap[t.Port]

		if !ok {
			m := http.Map{
				Source:   "$ssl_preread_server_name",
				Variable: "$dest" + fmt.Sprint(t.Port),
				Parameters: []http.MapParameter{
					{
						Value:  t.Hostname,
						Result: "unix:/var/lib/nginx/" + t.Hostname + fmt.Sprint(t.Port) + ".sock",
					},
				},
			}
			maps = append(maps, &m)
			portsToMap[t.Port] = &m
		} else {
			streamMap.Parameters = append(streamMap.Parameters, http.MapParameter{
				Value:  t.Hostname,
				Result: "unix:/var/lib/nginx/" + t.Hostname + fmt.Sprint(t.Port) + ".sock",
			})
		}
	}

	return maps
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
	// header modifiers, we need to create a map parameter regex for any string value.
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
