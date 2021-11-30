package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	flag "github.com/spf13/pflag"
)

const (
	errTmpl = "failed validation - flag: '--%s' reason: '%s'\n"
)

type Validator func(*flag.FlagSet) error
type ValidatorContext struct {
	Key string
	V   Validator
}

func GatewayControllerParam(domain string, namespace string) ValidatorContext {
	name := "gateway-ctlr-name"
	return ValidatorContext{
		name,
		func(flagset *flag.FlagSet) error {
			// FIXME(yacobucci) this does not provide the same regex validation as
			// GatewayClass.ControllerName. provide equal and then specific validation
			param, err := flagset.GetString(name)
			if err != nil {
				return err
			}

			if len(param) == 0 {
				return errors.New("flag must be set")
			}

			fields := strings.Split(param, "/")
			l := len(fields)
			if l != 3 {
				return errors.New("unsupported path length, must be form DOMAIN/NAMESPACE/NAME")
			}

			for i := len(fields); i > 0; i-- {
				switch i {
				case 3:
					if fields[0] != domain {
						return fmt.Errorf("invalid domain: %s", fields[0])
					}
					fields = fields[1:]
				case 2:
					if fields[0] != namespace {
						return fmt.Errorf("cross namespace unsupported: %s", fields[0])
					}
					fields = fields[1:]
				case 1:
					if fields[0] == "" {
						return errors.New("must provide a name")
					}
				}
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
