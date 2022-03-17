package file

import "testing"

func TestGetPathForServerConfig(t *testing.T) {
	expected := "/etc/nginx/conf.d/test.example.com.conf"

	result := getPathForServerConfig("test.example.com")
	if result != expected {
		t.Errorf("getPathForServerConfig() returned %q but expected %q", result, expected)
	}
}
