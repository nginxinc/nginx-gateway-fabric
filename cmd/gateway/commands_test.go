package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
)

type flagTestCase struct {
	name              string
	expectedErrPrefix string
	args              []string
	wantErr           bool
}

func testFlag(t *testing.T, cmd *cobra.Command, test flagTestCase) {
	t.Helper()
	g := NewWithT(t)
	// discard any output generated by cobra
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	// override RunE to avoid executing the command
	cmd.RunE = func(_ *cobra.Command, _ []string) error {
		return nil
	}

	cmd.SetArgs(test.args)
	err := cmd.Execute()

	if test.wantErr {
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(HavePrefix(test.expectedErrPrefix))
	} else {
		g.Expect(err).NotTo(HaveOccurred())
	}
}

func TestRootCmd(t *testing.T) {
	t.Parallel()
	testCase := flagTestCase{
		name:    "no flags",
		args:    nil,
		wantErr: false,
	}

	testFlag(t, createRootCommand(), testCase)
}

func TestCommonFlagsValidation(t *testing.T) {
	t.Parallel()
	tests := []flagTestCase{
		{
			name: "valid flags",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway",
				"--gatewayclass=nginx",
			},
			wantErr: false,
		},
		{
			name: "gateway-ctlr-name is not set",
			args: []string{
				"--gatewayclass=nginx",
			},
			wantErr:           true,
			expectedErrPrefix: `required flag(s) "gateway-ctlr-name" not set`,
		},
		{
			name: "gateway-ctlr-name is set to empty string",
			args: []string{
				"--gateway-ctlr-name=",
				"--gatewayclass=nginx",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--gateway-ctlr-name" flag: must be set`,
		},
		{
			name: "gateway-ctlr-name is invalid",
			args: []string{
				"--gateway-ctlr-name=nginx-gateway",
				"--gatewayclass=nginx",
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "nginx-gateway" for "--gateway-ctlr-name" flag: invalid format; ` +
				"must be DOMAIN/PATH",
		},
		{
			name: "gatewayclass is not set",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway",
			},
			wantErr:           true,
			expectedErrPrefix: `required flag(s) "gatewayclass" not set`,
		},
		{
			name: "gatewayclass is set to empty string",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway",
				"--gatewayclass=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--gatewayclass" flag: must be set`,
		},
		{
			name: "gatewayclass is invalid",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway",
				"--gatewayclass=@",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "@" for "--gatewayclass" flag: invalid format`,
		},
	}

	for _, test := range tests {
		t.Run(test.name+"_static_mode", func(t *testing.T) {
			t.Parallel()
			testFlag(t, createStaticModeCommand(), test)
		})
		t.Run(test.name+"_provisioner_mode", func(t *testing.T) {
			t.Parallel()
			testFlag(t, createProvisionerModeCommand(), test)
		})
	}
}

func TestStaticModeCmdFlagValidation(t *testing.T) {
	t.Parallel()
	tests := []flagTestCase{
		{
			name: "valid flags",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway", // common and required flag
				"--gatewayclass=nginx",                                // common and required flag
				"--gateway=nginx-gateway/nginx",
				"--config=nginx-gateway-config",
				"--service=nginx-gateway",
				"--update-gatewayclass-status=true",
				"--metrics-port=9114",
				"--metrics-disable",
				"--metrics-secure-serving",
				"--health-port=8081",
				"--health-disable",
				"--leader-election-lock-name=my-lock",
				"--leader-election-disable=false",
				"--nginx-plus",
				"--usage-report-secret=my-secret",
				"--usage-report-endpoint=example.com",
				"--usage-report-resolver=resolver.com",
				"--usage-report-ca-secret=ca-secret",
				"--usage-report-client-ssl-secret=client-secret",
				"--snippets-filters",
			},
			wantErr: false,
		},
		{
			name: "valid flags, non-required not set",
			args: []string{
				"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway", // common and required flag
				"--gatewayclass=nginx",                                // common and required flag,
			},
			wantErr: false,
		},
		{
			name: "gateway is set to empty string",
			args: []string{
				"--gateway=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--gateway" flag: must be set`,
		},
		{
			name: "gateway is invalid",
			args: []string{
				"--gateway=nginx-gateway", // no namespace
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "nginx-gateway" for "--gateway" flag: invalid format; ` +
				"must be NAMESPACE/NAME",
		},
		{
			name: "config is set to empty string",
			args: []string{
				"--config=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "-c, --config" flag: must be set`,
		},
		{
			name: "config is set to invalid string",
			args: []string{
				"--config=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "-c, --config" flag: invalid format`,
		},
		{
			name: "service is set to empty string",
			args: []string{
				"--service=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--service" flag: must be set`,
		},
		{
			name: "service is set to invalid string",
			args: []string{
				"--service=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "--service" flag: invalid format`,
		},
		{
			name: "update-gatewayclass-status is set to empty string",
			args: []string{
				"--update-gatewayclass-status=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--update-gatewayclass-status" flag: strconv.ParseBool`,
		},
		{
			name: "update-gatewayclass-status is invalid",
			args: []string{
				"--update-gatewayclass-status=invalid", // not a boolean
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "invalid" for "--update-gatewayclass-status" flag: strconv.ParseBool`,
		},
		{
			name: "metrics-port is invalid type",
			args: []string{
				"--metrics-port=invalid", // not an int
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "invalid" for "--metrics-port" flag: failed to parse int value:` +
				` strconv.ParseInt: parsing "invalid": invalid syntax`,
		},
		{
			name: "metrics-port is outside of range",
			args: []string{
				"--metrics-port=999", // outside of range
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "999" for "--metrics-port" flag:` +
				` port outside of valid port range [1024 - 65535]: 999`,
		},
		{
			name: "metrics-disable is not a bool",
			args: []string{
				"--metrics-disable=999", // not a bool
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "999" for "--metrics-disable" flag: strconv.ParseBool:` +
				` parsing "999": invalid syntax`,
		},
		{
			name: "metrics-secure-serving is not a bool",
			args: []string{
				"--metrics-secure-serving=999", // not a bool
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "999" for "--metrics-secure-serving" flag: strconv.ParseBool:` +
				` parsing "999": invalid syntax`,
		},
		{
			name: "health-port is invalid type",
			args: []string{
				"--health-port=invalid", // not an int
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "invalid" for "--health-port" flag: failed to parse int value:` +
				` strconv.ParseInt: parsing "invalid": invalid syntax`,
		},
		{
			name: "health-port is outside of range",
			args: []string{
				"--health-port=999", // outside of range
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "999" for "--health-port" flag:` +
				` port outside of valid port range [1024 - 65535]: 999`,
		},
		{
			name: "health-disable is not a bool",
			args: []string{
				"--health-disable=999", // not a bool
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "999" for "--health-disable" flag: strconv.ParseBool:` +
				` parsing "999": invalid syntax`,
		},
		{
			name: "leader-election-lock-name is set to invalid string",
			args: []string{
				"--leader-election-lock-name=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "--leader-election-lock-name" flag: invalid format`,
		},
		{
			name: "leader-election-disable is set to empty string",
			args: []string{
				"--leader-election-disable=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--leader-election-disable" flag: strconv.ParseBool`,
		},
		{
			name: "usage-report-secret is set to empty string",
			args: []string{
				"--usage-report-secret=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--usage-report-secret" flag: must be set`,
		},
		{
			name: "usage-report-secret is invalid",
			args: []string{
				"--usage-report-secret=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "--usage-report-secret" flag: invalid format: `,
		},
		{
			name: "usage-report-endpoint is set to empty string",
			args: []string{
				"--usage-report-endpoint=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--usage-report-endpoint" flag: must be set`,
		},
		{
			name: "usage-report-endpoint is an invalid endpoint",
			args: []string{
				"--usage-report-endpoint=$*(invalid)",
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "$*(invalid)" for "--usage-report-endpoint" flag: ` +
				`"$*(invalid)" must be a domain name or IP address with optional port`,
		},
		{
			name: "usage-report-resolver is set to empty string",
			args: []string{
				"--usage-report-resolver=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--usage-report-resolver" flag: must be set`,
		},
		{
			name: "usage-report-resolver is an invalid endpoint",
			args: []string{
				"--usage-report-resolver=$*(invalid)",
			},
			wantErr: true,
			expectedErrPrefix: `invalid argument "$*(invalid)" for "--usage-report-resolver" flag: ` +
				`"$*(invalid)" must be a domain name or IP address with optional port`,
		},
		{
			name: "usage-report-ca-secret is set to empty string",
			args: []string{
				"--usage-report-ca-secret=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--usage-report-ca-secret" flag: must be set`,
		},
		{
			name: "usage-report-ca-secret is invalid",
			args: []string{
				"--usage-report-ca-secret=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "--usage-report-ca-secret" flag: invalid format: `,
		},
		{
			name: "usage-report-client-ssl-secret is set to empty string",
			args: []string{
				"--usage-report-client-ssl-secret=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--usage-report-client-ssl-secret" flag: must be set`,
		},
		{
			name: "usage-report-client-ssl-secret is invalid",
			args: []string{
				"--usage-report-client-ssl-secret=!@#$",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "!@#$" for "--usage-report-client-ssl-secret" flag: invalid format: `,
		},
		{
			name: "snippets-filters is not a bool",
			expectedErrPrefix: `invalid argument "not-a-bool" for "--snippets-filters" flag: strconv.ParseBool:` +
				` parsing "not-a-bool": invalid syntax`,
			args: []string{
				"--snippets-filters=not-a-bool",
			},
			wantErr: true,
		},
	}

	// common flags validation is tested separately

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cmd := createStaticModeCommand()
			testFlag(t, cmd, test)
		})
	}
}

func TestProvisionerModeCmdFlagValidation(t *testing.T) {
	t.Parallel()
	testCase := flagTestCase{
		name: "valid flags",
		args: []string{
			"--gateway-ctlr-name=gateway.nginx.org/nginx-gateway", // common and required flag
			"--gatewayclass=nginx",                                // common and required flag
		},
		wantErr: false,
	}

	// common flags validation is tested separately

	testFlag(t, createProvisionerModeCommand(), testCase)
}

func TestSleepCmdFlagValidation(t *testing.T) {
	t.Parallel()
	tests := []flagTestCase{
		{
			name: "valid flags",
			args: []string{
				"--duration=1s",
			},
			wantErr: false,
		},
		{
			name:    "omitted flags",
			args:    nil,
			wantErr: false,
		},
		{
			name: "duration is set to empty string",
			args: []string{
				"--duration=",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "" for "--duration" flag: time: invalid duration ""`,
		},
		{
			name: "duration is invalid",
			args: []string{
				"--duration=invalid",
			},
			wantErr:           true,
			expectedErrPrefix: `invalid argument "invalid" for "--duration" flag: time: invalid duration "invalid"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cmd := createSleepCommand()
			testFlag(t, cmd, test)
		})
	}
}

func TestCopyCmdFlagValidation(t *testing.T) {
	t.Parallel()
	tests := []flagTestCase{
		{
			name: "valid flags",
			args: []string{
				"--source=/my/file",
				"--destination=dest/file",
			},
			wantErr: false,
		},
		{
			name:    "omitted flags",
			args:    nil,
			wantErr: false,
		},
		{
			name: "source set without destination",
			args: []string{
				"--source=/my/file",
			},
			wantErr: true,
			expectedErrPrefix: "if any flags in the group [source destination] are set they must all be set; " +
				"missing [destination]",
		},
		{
			name: "destination set without source",
			args: []string{
				"--destination=/dest/file",
			},
			wantErr: true,
			expectedErrPrefix: "if any flags in the group [source destination] are set they must all be set; " +
				"missing [source]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cmd := createCopyCommand()
			testFlag(t, cmd, test)
		})
	}
}

func TestCopyFile(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	src, err := os.CreateTemp(os.TempDir(), "testfile")
	g.Expect(err).ToNot(HaveOccurred())
	defer os.Remove(src.Name())

	dest, err := os.MkdirTemp(os.TempDir(), "testdir")
	g.Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dest)

	g.Expect(copyFile(src.Name(), dest)).To(Succeed())
	_, err = os.Stat(filepath.Join(dest, filepath.Base(src.Name())))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestParseFlags(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	flagSet := pflag.NewFlagSet("flagSet", 0)
	// set SortFlags to false for testing purposes so when parseFlags loops over the flagSet it
	// goes off of primordial order.
	flagSet.SortFlags = false

	var boolFlagTrue bool
	flagSet.BoolVar(
		&boolFlagTrue,
		"boolFlagTrue",
		true,
		"boolean true test flag",
	)

	var boolFlagFalse bool
	flagSet.BoolVar(
		&boolFlagFalse,
		"boolFlagFalse",
		false,
		"boolean false test flag",
	)

	customIntFlagDefault := intValidatingValue{
		validator: validatePort,
		value:     8080,
	}
	flagSet.Var(
		&customIntFlagDefault,
		"customIntFlagDefault",
		"default custom int test flag",
	)

	customIntFlagUserDefined := intValidatingValue{
		validator: validatePort,
		value:     8080,
	}
	flagSet.Var(
		&customIntFlagUserDefined,
		"customIntFlagUserDefined",
		"user defined custom int test flag",
	)
	err := flagSet.Set("customIntFlagUserDefined", "8081")
	g.Expect(err).To(Not(HaveOccurred()))

	customStringFlagDefault := stringValidatingValue{
		validator: validateResourceName,
		value:     "default-custom-string-test-flag",
	}
	flagSet.Var(
		&customStringFlagDefault,
		"customStringFlagDefault",
		"default custom string test flag",
	)

	customStringFlagUserDefined := stringValidatingValue{
		validator: validateResourceName,
		value:     "user-defined-custom-string-test-flag",
	}
	flagSet.Var(
		&customStringFlagUserDefined,
		"customStringFlagUserDefined",
		"user defined custom string test flag",
	)
	err = flagSet.Set("customStringFlagUserDefined", "changed-test-flag-value")
	g.Expect(err).To(Not(HaveOccurred()))

	customStringFlagNoDefaultValueUnset := namespacedNameValue{
		value: types.NamespacedName{},
	}
	flagSet.Var(
		&customStringFlagNoDefaultValueUnset,
		"customStringFlagNoDefaultValueUnset",
		"no default value custom string test flag",
	)

	customStringFlagNoDefaultValueUserDefined := namespacedNameValue{
		value: types.NamespacedName{},
	}
	flagSet.Var(
		&customStringFlagNoDefaultValueUserDefined,
		"customStringFlagNoDefaultValueUserDefined",
		"no default value but with user defined namespacedName test flag",
	)
	userDefinedNamespacedName := types.NamespacedName{
		Namespace: "changed-namespace",
		Name:      "changed-name",
	}
	err = flagSet.Set("customStringFlagNoDefaultValueUserDefined", userDefinedNamespacedName.String())
	g.Expect(err).To(Not(HaveOccurred()))

	expectedKeys := []string{
		"boolFlagTrue",
		"boolFlagFalse",

		"customIntFlagDefault",
		"customIntFlagUserDefined",

		"customStringFlagDefault",
		"customStringFlagUserDefined",

		"customStringFlagNoDefaultValueUnset",
		"customStringFlagNoDefaultValueUserDefined",
	}
	expectedValues := []string{
		"true",
		"false",

		"default",
		"user-defined",

		"default",
		"user-defined",

		"default",
		"user-defined",
	}

	flagKeys, flagValues := parseFlags(flagSet)

	g.Expect(flagKeys).Should(Equal(expectedKeys))
	g.Expect(flagValues).Should(Equal(expectedValues))
}

func TestGetBuildInfo(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	commitHash, commitTime, dirtyBuild := getBuildInfo()

	g.Expect(commitHash).To(Not(BeEmpty()))
	g.Expect(commitTime).To(Not(BeEmpty()))
	g.Expect(dirtyBuild).To(Not(BeEmpty()))

	g.Expect(commitHash).To(Not(Equal("unknown")))
	g.Expect(commitTime).To(Not(Equal("unknown")))
	g.Expect(dirtyBuild).To(Not(Equal("unknown")))
}
