package telemetry_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

var _ = Describe("Job", func() {
	var (
		job           *telemetry.Job
		exporter      *telemetryfakes.FakeExporter
		dataCollector *telemetryfakes.FakeDataCollector
	)
	const timeout = 10 * time.Second

	BeforeEach(func() {
		exporter = &telemetryfakes.FakeExporter{}
		dataCollector = &telemetryfakes.FakeDataCollector{}
		job = telemetry.NewJob(telemetry.JobConfig{
			Exporter:      exporter,
			Logger:        zap.New(),
			Period:        1 * time.Millisecond, // 1ms is much smaller than timeout so the Job should report a few times
			DataCollector: dataCollector,
		})
	})

	DescribeTable(
		"Job runs with a few reports without any errors",
		func(exporterError error) {
			// The fact that exporter return an error must not affect how many times the Job makes a report.
			exporter.ExportReturns(exporterError)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			errCh := make(chan error)
			go func() {
				errCh <- job.Start(ctx)
				close(errCh)
			}()

			const minReports = 2 // ensure that the Job reports more than once: it doesn't exit after the first report

			Eventually(exporter.ExportCallCount).Should(BeNumerically(">=", minReports))
			for i := 0; i < minReports; i++ {
				_, data := exporter.ExportArgsForCall(i)
				Expect(data).To(Equal(telemetry.Data{}))
			}

			cancel()
			Eventually(errCh).Should(Receive(BeNil()))
			Eventually(errCh).Should(BeClosed())
		},
		Entry("Job runs with Exporter not returning errors", nil),
		Entry("Job runs with Exporter returning an error", errors.New("some error")),
	)
})
