package controller

import (
	_ "sigs.k8s.io/controller-runtime/pkg/client"  // used below to generate a fake
	_ "sigs.k8s.io/controller-runtime/pkg/manager" // used below to generate a fake
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate  sigs.k8s.io/controller-runtime/pkg/manager.Manager

//counterfeiter:generate  sigs.k8s.io/controller-runtime/pkg/client.FieldIndexer
