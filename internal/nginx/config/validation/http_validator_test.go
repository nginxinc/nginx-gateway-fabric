package validation

import "testing"

func TestValidateHostnameInServer(t *testing.T) {
	validator := HTTPValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidateHostnameInServer,
		"",
		"example.com")
	testInvalidValuesForSimpleValidator(t, validator.ValidateHostnameInServer,
		`\`,
		`"`)
}
