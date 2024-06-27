package config

import (
	"strings"
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/shared"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
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

func createStreamMaps(conf dataplane.Configuration) []shared.Map {
	portsToMap := make(map[int32]shared.Map)

	for _, server := range conf.TLSPassthroughServers {
		streamMap, portInUse := portsToMap[server.Port]

		mapParam := shared.MapParameter{
			Value:  server.Hostname,
			Result: getSocketNameTLS(server.Port, server.Hostname),
		}

		if !portInUse {
			m := shared.Map{
				Source:   "$ssl_preread_server_name",
				Variable: getTLSPassthroughVarName(server.Port),
				Parameters: []shared.MapParameter{
					mapParam,
				},
				UseHostnames: true,
			}
			portsToMap[server.Port] = m
		} else {
			streamMap.Parameters = append(streamMap.Parameters, mapParam)
			portsToMap[server.Port] = streamMap
		}
	}

	for _, server := range conf.SSLServers {
		streamMap, portInUse := portsToMap[server.Port]

		hostname := server.Hostname

		if server.IsDefault {
			hostname = "default"
		}

		if portInUse {
			streamMap.Parameters = append(streamMap.Parameters, shared.MapParameter{
				Value:  hostname,
				Result: getSocketNameHTTPS(server.Port),
			})
			portsToMap[server.Port] = streamMap
		}
	}

	maps := make([]shared.Map, 0, len(portsToMap))

	for _, m := range portsToMap {
		maps = append(maps, m)
	}

	return maps
}

func buildAddHeaderMaps(servers []dataplane.VirtualServer) []shared.Map {
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

	maps := make([]shared.Map, 0, len(addHeaderNames))
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

func createAddHeadersMap(name string) shared.Map {
	underscoreName := convertStringToSafeVariableName(name)
	httpVarSource := "${http_" + underscoreName + "}"
	mapVarName := generateAddHeaderMapVariableName(name)
	params := []shared.MapParameter{
		{
			Value:  "default",
			Result: "''",
		},
		{
			Value:  anyStringFmt,
			Result: httpVarSource + ",",
		},
	}
	return shared.Map{
		Source:     httpVarSource,
		Variable:   "$" + mapVarName,
		Parameters: params,
	}
}
