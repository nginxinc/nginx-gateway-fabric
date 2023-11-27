package static

import (
	"context"

	"github.com/go-logr/logr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type errorHandler struct {
	logger logr.Logger
}

func (h errorHandler) Handle(err error) {
	h.logger.Error(err, "error in telemetry")
}

type telemetryReporter struct {
	k8sClient client.Client
	logger    logr.Logger
}

func (r *telemetryReporter) Start(ctx context.Context) error {
	// runs once in a POC.

	// telemetry endpoint configuration
	var (
		// /otel-collector/README.md deploys a collector with the endpoint below
		// no TLS or auth headers are needed for that collector
		endpoint = "my-opentelemetry-collector.default.svc.cluster.local:4317"
		secure   = false               // use TLS or not.
		headers  = map[string]string{} // allows t add headers to a GRPC connection. (e.g. authentication)
	)

	// gather telemetry data
	var ns v1.Namespace
	err := r.k8sClient.Get(ctx, client.ObjectKey{Name: "kube-system"}, &ns)
	if err != nil {
		return err
	}

	// cluster ID (UUID of kube-system namespace)
	clusterID := string(ns.UID)
	// NGF version
	ngfVersion := "product-tel-prototype-0.0.1"

	// Step 1
	// configure OTel global options
	otel.SetLogger(r.logger)
	otel.SetErrorHandler(errorHandler{
		logger: r.logger,
	})

	// Step 2
	// create a resource.
	// the resource represents the entity producing telemetry data.
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceName("NGF"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return err
	}

	// Step 3
	// Create exporters

	// stdout exporter will not be needed in production.
	stdoutExporter, err := newStdoutExporter()
	if err != nil {
		return err
	}

	// this creates an exporter with a GRPC connection to the collector
	// however, it will not establish a network connection until we start sending data.
	// see the comment inside newOTLPExporter()
	otelExporter, err := newOTLPExporter(ctx, endpoint, secure, headers)
	if err != nil {
		return err
	}

	// Step 4
	// Create provider
	// Batchers batch exporting multiple spans together for performance.
	// We don't need that, since we only need to send one span.
	provider := sdktrace.NewTracerProvider(
		//sdktrace.WithBatcher(stdoutExporter),
		//sdktrace.WithBatcher(otelExporter),
		sdktrace.WithSyncer(stdoutExporter), // prints to stdout
		sdktrace.WithSyncer(otelExporter),   // sends to a collector
		sdktrace.WithResource(res),
	)

	// Step 5
	// Create tracer
	tracer := provider.Tracer("product-telemetry")

	// Step 6
	// create span
	_, span := tracer.Start(ctx, "report")

	span.SetAttributes(
		attribute.String("clusterId", clusterID),
		attribute.String("ngfVersion", ngfVersion),
	)

	// Step 7
	// End span (send a trace).
	// if sdktrace.WithSyncer is used, this will block until the data is sent.
	// if sdktrace.WithBatcher is used, this will not block. However, the data will not be sent immediately.
	// Any errors will be passed to the errorHandler configured above.
	span.End()

	return nil
}

func newStdoutExporter() (*stdouttrace.Exporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

func newOTLPExporter(
	ctx context.Context,
	endpoint string,
	secure bool,
	headers map[string]string,
) (*otlptrace.Exporter, error) {
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithHeaders(headers),
		// Uncomment the block bellow to make sure a connection to the endpoint is established before otlptrace.New() returns.
		// Not recommended. See https://github.com/grpc/grpc-go/blob/master/Documentation/anti-patterns.md#dialing-in-grpc
		//otlptracegrpc.WithDialOption(
		//	grpc.WithBlock(),
		//),
	}

	if !secure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	traceClient := otlptracegrpc.NewClient(options...)

	return otlptrace.New(ctx, traceClient)
}
