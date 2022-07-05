package telemetry

import (
	"context"
	grpc_otel "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	jaegerPropagator "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel/attribute"
	jaegerExporter "go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// NewJaegerTracerProvider initializes a new Open Telemetry tracer provider for Jaeger.
// 	service: Describes the service that will be exporting traces into Jaeger. Usually contains the service name.
//	url: Contains the endpoint where to publish traces to. For Jaeger, it's the collector's endpoint.
//	environment: Used to identify the environment that a certain service is publishing traces from. Defaults to "development".
func NewJaegerTracerProvider(service, url, environment string) (trace.TracerProvider, error) {
	exporter, err := jaegerExporter.New(jaegerExporter.WithCollectorEndpoint(jaegerExporter.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	if environment == "" {
		environment = "development"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(res),
	)
	return tp, nil
}

// NewJaegerPropagator initializes a new Open Telemetry traces propagator for Jaeger.
func NewJaegerPropagator() jaegerPropagator.Jaeger {
	return jaegerPropagator.Jaeger{}
}

// NewSpan initializes a new span from the given context.
func NewSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// NewClientInterceptor initializes a new client interceptor for gRPC using the given propagator and tracer provider.
func NewClientInterceptor(p propagation.TextMapPropagator, tp trace.TracerProvider) (grpc.UnaryClientInterceptor, grpc.StreamClientInterceptor) {
	return grpc_otel.UnaryClientInterceptor(
			grpc_otel.WithPropagators(p),
			grpc_otel.WithTracerProvider(tp),
		),
		grpc_otel.StreamClientInterceptor(
			grpc_otel.WithPropagators(p),
			grpc_otel.WithTracerProvider(tp),
		)
}

// NewServerInterceptor initializes a new server interceptor for gRPC using the given propagator and tracer provider.
func NewServerInterceptor(p propagation.TextMapPropagator, tp trace.TracerProvider) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	return grpc_otel.UnaryServerInterceptor(
			grpc_otel.WithPropagators(p),
			grpc_otel.WithTracerProvider(tp),
		),
		grpc_otel.StreamServerInterceptor(
			grpc_otel.WithPropagators(p),
			grpc_otel.WithTracerProvider(tp),
		)
}
