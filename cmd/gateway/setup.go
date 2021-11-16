package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
)

type Validator func() (bool, error)

func GatewayControllerParam(required bool, domain string, namespace string, param string) Validator {
	return func() (bool, error) {
		if required && len(param) == 0 {
			return false, errors.New("gateway-ctlr-name must have a value")
		}

		fields := strings.Split(param, "/")
		l := len(fields)
		if l > 3 || l < 3 {
			return false, fmt.Errorf("unsupported path length, must be form DOMAIN/NAMESPACE/NAME")
		}

		for i := len(fields); i > 0; i-- {
			switch i {
			case 3:
				if fields[0] != domain {
					return false, fmt.Errorf("invalid domain: %s - %s", domain, param)
				}
				fields = fields[1:]
			case 2:
				if fields[0] != namespace {
					return false, fmt.Errorf("cross namespace unsupported: %s - %s", namespace, param)
				}
				fields = fields[1:]
			case 1:
				if fields[0] == "" {
					return false, fmt.Errorf("must provide a name: %s", param)
				}
			}
		}

		return true, nil
	}
}

func ValidateArguments(logger logr.Logger, validators ...Validator) bool {
	valid := true
	for _, v := range validators {
		if r, err := v(); !r {
			logger.Error(err, "failed validation")
			if valid {
				valid = !valid
			}
		}
	}
	return valid
}
