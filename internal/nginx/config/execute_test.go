package config

import (
	"testing"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/http"
)

func TestExecute(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("execute() panicked with %v", r)
		}
	}()

	bytes := execute(serversTemplate, []http.Server{})
	if len(bytes) == 0 {
		t.Error("template.execute() did not generate anything")
	}
}

func TestExecutePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("template.execute() did not panic")
		}
	}()

	_ = execute(serversTemplate, "not-correct-data")
}
