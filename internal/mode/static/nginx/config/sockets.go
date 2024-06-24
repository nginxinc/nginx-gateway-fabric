package config

import (
	"fmt"
)

func getSocketName(port int32, hostname string) string {
	return fmt.Sprintf("unix:/var/run/nginx/%s%d.sock", hostname, port)
}

func getVariableName(port int32) string {
	return fmt.Sprintf("dest%d", port)
}
