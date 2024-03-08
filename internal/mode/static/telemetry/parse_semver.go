package telemetry

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

// parseSemver takes a string and turns it into a semver format.
func parseSemver(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")

	if s == "" {
		return "", errors.New("string cannot be empty")
	}

	parts := strings.SplitN(s, ".", 3)

	if _, err := strconv.Atoi(parts[0]); err != nil {
		return "", errors.New("string must have a number as the major version")
	}

	if len(parts) == 1 || parts[1] == "" {
		return "", errors.New("string must have at least a major and minor version specified")
	}

	// in the edge case where the kubeletVersion is missing the patch version and has trailing characters which include
	// '.' e.g. "v1.27-gke.1067004"
	if _, err := strconv.Atoi(parts[1]); err != nil && len(parts) == 3 {
		parts[1] = parts[1] + parts[2]
		parts = parts[:len(parts)-1]
	}

	lastString := parts[len(parts)-1]

	for index := range lastString {
		// cut off trailing characters after the patch version.
		// e.g. if kubeletVersion = "1.27.4+500050039", will return "4"
		if !unicode.IsDigit(rune(lastString[index])) {
			parts[len(parts)-1] = lastString[:index]
			break
		}
	}

	if len(parts) == 2 {
		parts = append(parts, "0")
	}

	// in the case where lastString was ""
	if parts[2] == "" {
		parts[2] = "0"
	}

	return strings.Join(parts, "."), nil
}
