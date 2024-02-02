package telemetry_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/runnables"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/telemetry/telemetryfakes"
)

var _ = Describe("Job", func() {
	var (
		cronjob         *runnables.CronJob
		exporter        *telemetryfakes.FakeExporter
		dataCollector   *telemetryfakes.FakeDataCollector
		healthCollector *telemetryfakes.FakeHealthChecker
		expData         telemetry.Data
		worker          func(context.Context)
		readyChannel    chan struct{}
	)
	const timeout = 10 * time.Second

	BeforeEach(func() {
		exporter = &telemetryfakes.FakeExporter{}
		dataCollector = &telemetryfakes.FakeDataCollector{}
		healthCollector = &telemetryfakes.FakeHealthChecker{}

		readyChannel = make(chan struct{})
		healthCollector.GetReadyIfClosedChannelReturns(readyChannel)

		worker = telemetry.CreateTelemetryJobWorker(zap.New(), exporter, dataCollector, healthCollector)

		cronjob = runnables.NewCronJob(runnables.CronJobConfig{
			Worker:       worker,
			Logger:       zap.New(),
			Period:       1 * time.Millisecond, // 1ms is much smaller than timeout so the Job should report a few times,
			JitterFactor: 10.0 / (24 * 60),     // added jitter is bound by jitterFactor * period
		})

		expData = telemetry.Data{
			ProjectMetadata: telemetry.ProjectMetadata{Name: "NGF", Version: "1.1"},
			NodeCount:       3,
			NGFResourceCounts: telemetry.NGFResourceCounts{
				Gateways:       1,
				GatewayClasses: 1,
				HTTPRoutes:     1,
				Secrets:        1,
				Services:       1,
				Endpoints:      1,
			},
		}
		dataCollector.CollectReturns(expData, nil)
	})

	DescribeTable(
		"Job runs with a few reports without any errors",
		func(exporterError error, sleep time.Duration) {
			// The fact that exporter return an error must not affect how many times the Job makes a report.
			exporter.ExportReturns(exporterError)

			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			errCh := make(chan error)
			go func() {
				errCh <- cronjob.Start(ctx)
				close(errCh)
			}()
			time.Sleep(sleep)
			close(readyChannel)

			const minReports = 2 // ensure that the Job reports more than once: it doesn't exit after the first report

			Eventually(exporter.ExportCallCount).Should(BeNumerically(">=", minReports))
			for i := 0; i < minReports; i++ {
				_, data := exporter.ExportArgsForCall(i)
				Expect(data).To(Equal(expData))
			}

			cancel()
			Eventually(errCh).Should(Receive(BeNil()))
			Eventually(errCh).Should(BeClosed())
		},
		Entry("Job runs with Exporter not returning errors", nil, time.Duration(0)),
		Entry("Job runs with Exporter returning an error", errors.New("some error"), time.Duration(0)),
		Entry("Job runs with extended time waiting for NGF Pod to be ready", nil, 500*time.Millisecond),
	)

	When("job context is canceled", func() {
		It("should gracefully exist", func() {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)

			errCh := make(chan error)
			go func() {
				errCh <- cronjob.Start(ctx)
				close(errCh)
			}()

			sleep := 500 * time.Millisecond
			time.Sleep(sleep)
			cancel()

			Eventually(errCh).Should(Receive())
			Eventually(errCh).Should(BeClosed())
		})
	})
})
