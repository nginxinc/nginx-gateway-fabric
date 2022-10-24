package config

import (
	"strings"
)

// NGINX Variable names cannot have hyphens.
// This function converts a hyphenated string to an underscored string.
func convertStringToSafeVariableName(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}
