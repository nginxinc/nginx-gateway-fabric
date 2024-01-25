package telemetry_test

import (
	. "github.com/onsi/ginkgo/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
)

var _ = Describe("Collector", func() {
	var (
		k8sClient client.Client
		_         telemetry.DataCollector
		version   string
	)
	BeforeEach(func() {
		version = "1.1"
		k8sClient = fake.NewFakeClient()
		_ = telemetry.NewDataCollector(telemetry.DataCollectorConfig{K8sClientReader: k8sClient, Version: version})
	})
})
