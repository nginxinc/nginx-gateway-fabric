package config

import (
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/shared"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func createIncludeExecuteResultsFromServers(servers []http.Server) []executeResult {
	uniqueIncludes := make(map[string][]byte)

	// deduplicate include files across servers and location
	for _, server := range servers {
		for _, include := range server.Includes {
			uniqueIncludes[include.Name] = include.Content
		}

		for _, loc := range server.Locations {
			for _, include := range loc.Includes {
				uniqueIncludes[include.Name] = include.Content
			}
		}
	}

	results := make([]executeResult, 0, len(uniqueIncludes))

	for filename, contents := range uniqueIncludes {
		results = append(
			results, executeResult{
				dest: filename,
				data: contents,
			},
		)
	}

	return results
}

func createIncludesFromPolicyGenerateResult(resFiles []policies.File) []shared.Include {
	if len(resFiles) == 0 {
		return nil
	}

	includes := make([]shared.Include, 0, len(resFiles))
	for _, file := range resFiles {
		includes = append(
			includes, shared.Include{
				Name:    includesFolder + "/" + file.Name,
				Content: file.Content,
			},
		)
	}

	return includes
}

func createIncludeFromSnippet(snippet dataplane.Snippet) shared.Include {
	return shared.Include{
		Name:    includesFolder + "/" + snippet.Name + ".conf",
		Content: []byte(snippet.Contents),
	}
}

func deduplicateIncludes(includes []shared.Include) []shared.Include {
	uniqueIncludes := make(map[string]shared.Include)
	for _, i := range includes {
		if _, ok := uniqueIncludes[i.Name]; !ok {
			uniqueIncludes[i.Name] = i
		}
	}

	results := make([]shared.Include, 0, len(uniqueIncludes))
	for _, i := range uniqueIncludes {
		results = append(results, i)
	}

	return results
}

func createIncludesFromLocationSnippetsFilters(filters []dataplane.SnippetsFilter) []shared.Include {
	if len(filters) == 0 {
		return nil
	}

	includes := make([]shared.Include, 0)

	if len(filters) > 0 {
		for _, f := range filters {
			if f.LocationSnippet != nil {
				includes = append(includes, createIncludeFromSnippet(*f.LocationSnippet))
			}
		}
	}

	return deduplicateIncludes(includes)
}

func createIncludesFromServerSnippetsFilters(server dataplane.VirtualServer) []shared.Include {
	if len(server.PathRules) == 0 {
		return nil
	}

	includes := make([]shared.Include, 0)

	for _, pr := range server.PathRules {
		for _, mr := range pr.MatchRules {
			if len(mr.Filters.SnippetsFilters) > 0 {
				for _, sf := range mr.Filters.SnippetsFilters {
					if sf.ServerSnippet != nil {
						includes = append(includes, createIncludeFromSnippet(*sf.ServerSnippet))
					}
				}
			}
		}
	}

	return deduplicateIncludes(includes)
}

func createIncludesFromSnippets(snippets []dataplane.Snippet) []shared.Include {
	if len(snippets) == 0 {
		return nil
	}

	includes := make([]shared.Include, 0)

	for _, s := range snippets {
		includes = append(includes, createIncludeFromSnippet(s))
	}

	return deduplicateIncludes(includes)
}

func createIncludeExecuteResults(includes []shared.Include) []executeResult {
	results := make([]executeResult, 0, len(includes))

	for _, inc := range includes {
		results = append(
			results, executeResult{
				dest: inc.Name,
				data: inc.Content,
			},
		)
	}

	return results
}
