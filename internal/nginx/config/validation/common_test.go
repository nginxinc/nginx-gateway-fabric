package validation

import (
	"testing"
)

func TestValidateEscapedString(t *testing.T) {
	validator := func(value string) error { return validateEscapedString(value, []string{"example"}) }

	testValidValuesForSimpleValidator(t, validator,
		`test`,
		`test test`,
		`\"`,
		`\\`)
	testInvalidValuesForSimpleValidator(t, validator,
		`\`,
		`test"test`)
}

func TestValidateEscapedStringNoVarExpansion(t *testing.T) {
	validator := func(value string) error { return validateEscapedStringNoVarExpansion(value, []string{"example"}) }

	testValidValuesForSimpleValidator(t, validator,
		`test`,
		`test test`,
		`\"`,
		`\\`)
	testInvalidValuesForSimpleValidator(t, validator,
		`\`,
		`test"test`,
		`$test`)
}
