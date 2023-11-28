//go:build conformance

/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package tests

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"
)

func TestConformance(t *testing.T) {
	g := NewWithT(t)
	cfg, err := config.GetConfig()
	g.Expect(err).To(BeNil())

	client, err := client.New(cfg, client.Options{})
	g.Expect(err).To(BeNil())

	g.Expect(v1alpha2.AddToScheme(client.Scheme())).To(Succeed())
	g.Expect(v1.AddToScheme(client.Scheme())).To(Succeed())
	g.Expect(v1beta1.AddToScheme(client.Scheme())).To(Succeed())

	supportedFeatures := suite.ParseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures := suite.ParseSupportedFeatures(*flags.ExemptFeatures)

	t.Logf(`Running conformance tests with %s GatewayClass\n cleanup: %t\n`+
		`debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]`,
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug,
		*flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures)

	expSuite, err := suite.NewExperimentalConformanceTestSuite(suite.ExperimentalConformanceOptions{
		Options: suite.Options{
			Client:                     client,
			GatewayClassName:           *flags.GatewayClassName,
			Debug:                      *flags.ShowDebug,
			CleanupBaseResources:       *flags.CleanupBaseResources,
			SupportedFeatures:          supportedFeatures,
			ExemptFeatures:             exemptFeatures,
			EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
		},
		Implementation: v1alpha1.Implementation{
			Organization: "nginxinc",
			Project:      "nginx-gateway-fabric",
			URL:          "https://github.com/nginxinc/nginx-gateway-fabric",
			Version:      *flags.ImplementationVersion,
			Contact: []string{
				"https://github.com/nginxinc/nginx-gateway-fabric/discussions/new/choose",
			},
		},
		ConformanceProfiles: sets.New(suite.HTTPConformanceProfileName),
	})
	g.Expect(err).To(Not(HaveOccurred()))

	expSuite.Setup(t)
	err = expSuite.Run(t, tests.ConformanceTests)
	g.Expect(err).To(Not(HaveOccurred()))

	report, err := expSuite.Report()
	g.Expect(err).To(Not(HaveOccurred()))

	yamlReport, err := yaml.Marshal(report)
	g.Expect(err).ToNot(HaveOccurred())

	f, err := os.Create(*flags.ReportOutput)
	g.Expect(err).ToNot(HaveOccurred())
	defer f.Close()

	_, err = f.WriteString("CONFORMANCE PROFILE\n")
	g.Expect(err).ToNot(HaveOccurred())

	_, err = f.Write(yamlReport)
	g.Expect(err).ToNot(HaveOccurred())
}
