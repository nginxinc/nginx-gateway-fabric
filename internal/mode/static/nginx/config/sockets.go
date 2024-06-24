package config

import (
	"fmt"
	"strings"
)

var forbiddenChars = map[string]string{":": "::", "*": ":s"}

// Swap forbidden characters treating ":" as an escape character
func swapCharacters(name string) string {
	for old, replace := range forbiddenChars {
		name = strings.Replace(name, old, replace, -1)
	}
	return name
}

func getSocketName(port int32, hostname string) string {
	newName := swapCharacters(hostname)
	return fmt.Sprintf("unix:/var/run/nginx/%s%d.sock", newName, port)
}

func getVariableName(port int32) string {
	return fmt.Sprintf("dest%d", port)
}
