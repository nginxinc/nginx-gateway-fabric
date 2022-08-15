package config

import (
	"testing"
	"text/template"
)

func TestExecuteForServer(t *testing.T) {
	executor := newTemplateExecutor()

	http := http{
		Servers: []server{
			{
				ServerName: "example.com",
				Locations: []location{
					{
						Path:      "/",
						ProxyPass: "http://example-upstream",
					},
				},
			},
		},
		Upstreams: []upstream{
			{
				Name: "example-upstream",
				Servers: []upstreamServer{
					{
						Address: "http://10.0.0.1:80",
					},
				},
			},
		},
	}

	cfg := executor.ExecuteForHTTP(http)
	// we only do a sanity check here.
	// the config generation logic is tested in the Generator tests.
	if len(cfg) == 0 {
		t.Error("ExecuteForHTTP() returned 0-length config")
	}
}

func TestNewTemplateExecutorPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("newTemplateExecutor() didn't panic")
		}
	}()

	httpTemplate = "{{ end }}" // invalid template
	newTemplateExecutor()
}

func TestExecuteForServerPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("ExecuteForHTTP() didn't panic")
		}
	}()

	tmpl, err := template.New("test").Parse("{{ .NonExistingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	executor := &templateExecutor{httpTemplate: tmpl}

	_ = executor.ExecuteForHTTP(http{})
}
