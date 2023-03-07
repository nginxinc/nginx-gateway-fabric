package validation

import (
	"fmt"
	"sort"
)

func validateInSupportedValues[T comparable](
	value T,
	supportedValues map[T]struct{},
) (valid bool, supportedValuesAsStrings []string) {
	if _, exist := supportedValues[value]; exist {
		return true, nil
	}

	return false, getSortedKeysAsString(supportedValues)
}

func getSortedKeysAsString[T comparable](m map[T]struct{}) []string {
	keysAsString := make([]string, 0, len(m))

	for k := range m {
		keysAsString = append(keysAsString, fmt.Sprint(k))
	}

	sort.Strings(keysAsString)

	return keysAsString
}
