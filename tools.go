//go:build tools
// +build tools

// This file just exists to ensure we download the tools we need for building
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

import (
	_ "github.com/ahmetb/gen-crd-api-reference-docs"
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
