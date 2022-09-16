package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation"

	// Adding a dummy import here to remind us to check the controllerNameRegex when we update the Gateway API version.
	_ "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	errTmpl             = "failed validation - flag: '--%s' reason: '%s'\n"
	controllerNameRegex = `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$`
)

type (
	Validator        func(*flag.FlagSet) error
	ValidatorContext struct {
		Key string
		V   Validator
	}
)

func GatewayControllerParam(domain string) ValidatorContext {
	name := "gateway-ctlr-name"
	return ValidatorContext{
		name,
		func(flagset *flag.FlagSet) error {
			param, err := flagset.GetString(name)
			if err != nil {
				return err
			}

			if len(param) == 0 {
				return errors.New("flag must be set")
			}

			fields := strings.Split(param, "/")
			l := len(fields)
			if l < 2 {
				return errors.New("invalid format; must be DOMAIN/PATH")
			}

			if fields[0] != domain {
				return fmt.Errorf("invalid domain: %s; domain must be: %s", fields[0], domain)
			}

			return validateControllerName(param)
		},
	}
}

func validateControllerName(name string) error {
	// Regex from: https://github.com/kubernetes-sigs/gateway-api/blob/547122f7f55ac0464685552898c560658fb40073/apis/v1alpha2/shared_types.go#L462
	re := regexp.MustCompile(controllerNameRegex)
	if !re.Match([]byte(name)) {
		return fmt.Errorf("invalid gateway controller name: %s; expected format is DOMAIN/PATH", name)
	}
	return nil
}

func GatewayClassParam() ValidatorContext {
	name := "gatewayclass"
	return ValidatorContext{
		name,
		func(flagset *flag.FlagSet) error {
			param, err := flagset.GetString(name)
			if err != nil {
				return err
			}

			if len(param) == 0 {
				return errors.New("flag must be set")
			}

			// used by Kubernetes to validate resource names
			messages := validation.IsDNS1123Subdomain(param)
			if len(messages) > 0 {
				msg := strings.Join(messages, "; ")
				return fmt.Errorf("invalid format: %s", msg)
			}

			return nil
		},
	}
}

func ValidateArguments(flagset *flag.FlagSet, validators ...ValidatorContext) []string {
	var msgs []string
	for _, v := range validators {
		if flagset.Lookup(v.Key) != nil {
			err := v.V(flagset)
			if err != nil {
				msgs = append(msgs, fmt.Sprintf(errTmpl, v.Key, err.Error()))
			}
		}
	}

	return msgs
}

func MustValidateArguments(flagset *flag.FlagSet, validators ...ValidatorContext) {
	msgs := ValidateArguments(flagset, validators...)
	if msgs != nil {
		for i := range msgs {
			fmt.Fprintf(os.Stderr, "%s", msgs[i])
		}
		fmt.Fprintln(os.Stderr, "")

		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		os.Exit(1)
	}
}
