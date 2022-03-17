package config

import (
	"testing"
	"text/template"
)

func TestExecuteForServer(t *testing.T) {
	executor := newTemplateExecutor()

	server := server{
		ServerName: "example.com",
		Locations: []location{
			{
				Path:      "/",
				ProxyPass: "http://10.0.0.1",
			},
		},
	}

	cfg := executor.ExecuteForServer(server)
	// we only do a sanity check here.
	// the config generation logic is tested in the Generator tests.
	if len(cfg) == 0 {
		t.Error("ExecuteForServer() returned 0-length config")
	}
}

func TestNewTemplateExecutorPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("newTemplateExecutor() didn't panic")
		}
	}()

	serverTemplate = "{{ end }}" // invalid template
	newTemplateExecutor()
}

func TestExecuteForServerPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("ExecuteForServer() didn't panic")
		}
	}()

	tmpl, err := template.New("test").Parse("{{ .NonExistingField }}")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	executor := &templateExecutor{serverTemplate: tmpl}

	_ = executor.ExecuteForServer(server{})
}
