package file

import "testing"

func TestGetPathForServerConfig(t *testing.T) {
	expected := "/etc/nginx/conf.d/test.example.com.conf"

	result := getPathForConfig("test.example.com")
	if result != expected {
		t.Errorf("getPathForConfig() returned %q but expected %q", result, expected)
	}
}
