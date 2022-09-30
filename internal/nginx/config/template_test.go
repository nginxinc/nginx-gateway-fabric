package config

import (
	"testing"
	"text/template"
)

func TestExecuteForHTTPServers(t *testing.T) {
	executor := newTemplateExecutor()

	servers := httpServers{
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
	}

	cfg := executor.ExecuteForHTTPServers(servers)
	// we only do a sanity check here.
	// the config generation logic is tested in the Generator tests.
	if len(cfg) == 0 {
		t.Error("ExecuteForHTTPServers() returned 0-length config")
	}
}

func TestExecuteForHTTPUpstreams(t *testing.T) {
	executor := newTemplateExecutor()

	upstreams := httpUpstreams{
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

	cfg := executor.ExecuteForHTTPUpstreams(upstreams)
	// we only do a sanity check here.
	// the config generation logic is tested in the Generator tests.
	if len(cfg) == 0 {
		t.Error("ExecuteForHTTPUpstreams() returned 0-length config")
	}
}

func TestNewTemplateExecutorPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("newTemplateExecutor() didn't panic")
		}
	}()

	httpServersTemplate = "{{ end }}" // invalid template
	newTemplateExecutor()
}

func TestExecuteForHTTPServersPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("ExecuteForHTTPServers() didn't panic")
		}
	}()

	tmpl, err := template.New("test").Parse("{{ .NonExistingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	executor := &templateExecutor{httpServersTemplate: tmpl}

	_ = executor.ExecuteForHTTPServers(httpServers{})
}

func TestExecuteForHTTPUpstreamsPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("ExecuteForHTTPUpstreams() didn't panic")
		}
	}()

	tmpl, err := template.New("test").Parse("{{ .NonExistingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	executor := &templateExecutor{httpUpstreamsTemplate: tmpl}

	_ = executor.ExecuteForHTTPUpstreams(httpUpstreams{})
}
