package validation

import (
	"math"
	"testing"
)

func TestValidateRedirectScheme(t *testing.T) {
	validator := HTTPRedirectValidator{}

	testValidValuesForSupportedValuesValidator(
		t,
		validator.ValidateRedirectScheme,
		"http",
		"https",
	)

	testInvalidValuesForSupportedValuesValidator(
		t,
		validator.ValidateRedirectScheme,
		supportedRedirectSchemes,
		"test",
	)
}

func TestValidateRedirectPort(t *testing.T) {
	validator := HTTPRedirectValidator{}

	testValidValuesForSimpleValidator(
		t,
		validator.ValidateRedirectPort,
		math.MinInt32,
		math.MaxInt32,
	)
}

func TestValidateRedirectStatusCode(t *testing.T) {
	validator := HTTPRedirectValidator{}

	testValidValuesForSupportedValuesValidator(
		t,
		validator.ValidateRedirectStatusCode,
		301,
		302)

	testInvalidValuesForSupportedValuesValidator(
		t,
		validator.ValidateRedirectStatusCode,
		supportedRedirectStatusCodes,
		404,
	)
}

func TestValidateHostname(t *testing.T) {
	validator := HTTPRedirectValidator{}

	testValidValuesForSimpleValidator(
		t,
		validator.ValidateHostname,
		"example.com",
	)

	testInvalidValuesForSimpleValidator(
		t,
		validator.ValidateHostname,
		"example.com$",
	)
}

func TestValidateRewritePath(t *testing.T) {
	validator := HTTPURLRewriteValidator{}

	testValidValuesForSimpleValidator(
		t,
		validator.ValidateRewritePath,
		"",
		"/path",
		"/longer/path",
		"/trailing/",
	)

	testInvalidValuesForSimpleValidator(
		t,
		validator.ValidateRewritePath,
		"path",
		"$path",
		"/path$",
	)
}

func TestValidateRequestHeaderName(t *testing.T) {
	validator := HTTPRequestHeaderValidator{}

	testValidValuesForSimpleValidator(
		t,
		validator.ValidateRequestHeaderName,
		"Content-Encoding",
		"MyBespokeHeader",
	)

	testInvalidValuesForSimpleValidator(t, validator.ValidateRequestHeaderName, "$Content-Encoding")
}

func TestValidateRequestHeaderValue(t *testing.T) {
	validator := HTTPRequestHeaderValidator{}

	testValidValuesForSimpleValidator(
		t,
		validator.ValidateRequestHeaderValue,
		"my-cookie-name",
		"ssl_(server_name}",
		"example/1234==",
		"1234:3456",
	)

	testInvalidValuesForSimpleValidator(
		t,
		validator.ValidateRequestHeaderValue,
		"$Content-Encoding",
		`"example"`,
	)
}
