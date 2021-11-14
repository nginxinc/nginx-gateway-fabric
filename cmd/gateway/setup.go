package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
)

type Validator func() (bool, error)

func GatewayControllerParam(required bool, namespace string, param string) Validator {
	return func() (bool, error) {
		if required && len(param) == 0 {
			return false, errors.New("gateway-ctlr-name must have a value")
		}

		fields := strings.Split(param, "/")
		l := len(fields)
		if l > 3 {
			return false, fmt.Errorf("unsupported path length")
		}

		switch len(fields) {
		case 3:
			if fields[0] != domain {
				return false, fmt.Errorf("invalid domain: %s", domain)
			}
			fields = fields[1:]
			fallthrough
		case 2:
			if fields[0] != namespace {
				return false, fmt.Errorf("cannot cross namespace references: %s", namespace)
			}
			fields = fields[1:]
			fallthrough
		case 1:
			if fields[0] == "" {
				return false, fmt.Errorf("must provide a name")
			}
		}

		return true, nil
	}
}

func validateArguments(logger logr.Logger, validators ...Validator) bool {
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
