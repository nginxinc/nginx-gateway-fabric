package config

import (
	"github.com/go-logr/logr"
)

type Config struct {
	GatewayCtlrName string
	Logger          logr.Logger
}
