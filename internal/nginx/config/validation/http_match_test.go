package validation

import (
	"testing"
)

func TestValidatePathInPrefixMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidatePathInPrefixMatch,
		"/",
		"/path",
		"/path/subpath-123")
	testInvalidValuesForSimpleValidator(t, validator.ValidatePathInPrefixMatch,
		"/ ",
		"/path{",
		"/path}",
		"/path;",
		"path",
		"",
		"/path$")
}

func TestValidateHeaderNameInMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidateHeaderNameInMatch,
		"header")
	testInvalidValuesForSimpleValidator(t, validator.ValidateHeaderNameInMatch,
		":",
		"")
}

func TestValidateHeaderValueInMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidateHeaderValueInMatch,
		"value")
	testInvalidValuesForSimpleValidator(t, validator.ValidateHeaderValueInMatch,
		":",
		"")
}

func TestValidateQueryParamNameInMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidateQueryParamNameInMatch,
		"param")
	testInvalidValuesForSimpleValidator(t, validator.ValidateQueryParamNameInMatch,
		"")
}

func TestValidateQueryParamValueInMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSimpleValidator(t, validator.ValidateQueryParamValueInMatch,
		"value")
	testInvalidValuesForSimpleValidator(t, validator.ValidateQueryParamValueInMatch,
		"")
}

func TestValidateMethodInMatch(t *testing.T) {
	validator := HTTPMatchValidator{}

	testValidValuesForSupportedValuesValidator(t, validator.ValidateMethodInMatch,
		"GET")
	testInvalidValuesForSupportedValuesValidator(t, validator.ValidateMethodInMatch, supportedMethods,
		"GOT",
		"TRACE")
}

func TestValidateCommonMatchPart(t *testing.T) {
	testValidValuesForSimpleValidator(t, validateCommonMatchPart,
		"test")
	testInvalidValuesForSimpleValidator(t, validateCommonMatchPart,
		"",
		" ",
		"$")
}
