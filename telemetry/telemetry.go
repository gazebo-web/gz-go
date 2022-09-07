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
	// Define where traces will be exported to.
	// This block defines the endpoint to collect traces.
	exporter, err := jaegerExporter.New(
		jaegerExporter.WithCollectorEndpoint(
			jaegerExporter.WithEndpoint(url),
		),
	)
	if err != nil {
		return nil, err
	}

	// Set a default environment if no environment is provided.
	if environment == "" {
		environment = "development"
	}

	// Define the metadata for every trace.
	res, err := resource.Merge(
		resource.Default(), // Use the default ones
		resource.NewWithAttributes(
			semconv.SchemaURL, // Define the schema version being used (Open Telemetry v1.10.0)
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// Define the sampling, what exporter should be used, and what metadata should be embedded on every trace.
	tp := tracesdk.NewTracerProvider(
		// TODO: Change sampling strategy once we start having too many requests to avoid overloading the system.
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(res),
	)
	return tp, nil
}

// NewJaegerPropagator initializes a new Open Telemetry traces propagator for Jaeger.
// The propagator serializes and deserializes Jaeger headers to/from a context.Context.
func NewJaegerPropagator() jaegerPropagator.Jaeger {
	return jaegerPropagator.Jaeger{}
}

// NewSpan initializes a new span from the given context.
// Span is the individual component of a trace. It represents a single named
// and timed operation of a workflow that is traced. A Tracer is used to
// create a Span, and it is then up to the operation the Span represents to
// properly end the Span when the operation itself ends.
// If no Span is currently set in ctx a NoOp span is returned instead.
func NewSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// NewChildSpan initializes a new child span.
func NewChildSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	span := NewSpan(ctx)
	if !span.IsRecording() {
		return ctx, span
	}
	return span.TracerProvider().Tracer("").Start(ctx, name)
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

// AppendDialOptions appends unary and stream interceptors using the given propagator and tracer provider.
// If either propagator or tracerProvider are nil, this function returns the given opts as they were provided.
func AppendDialOptions(opts []grpc.DialOption, propagator propagation.TextMapPropagator, tracerProvider trace.TracerProvider) []grpc.DialOption {
	if propagator != nil && tracerProvider != nil {
		unaryInterceptor, streamInterceptor := NewClientInterceptor(propagator, tracerProvider)
		opts = append(opts, grpc.WithUnaryInterceptor(unaryInterceptor), grpc.WithStreamInterceptor(streamInterceptor))
	}
	return opts
}

// AppendServerInterceptors appends unary and stream interceptors using the given propagator and tracer provider.
// If either propagator or tracerProvider are nil, this function returns the given streams and unaries as they were provided.
func AppendServerInterceptors(unaries []grpc.UnaryServerInterceptor, streams []grpc.StreamServerInterceptor,
	p propagation.TextMapPropagator, tp trace.TracerProvider) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	if p != nil && tp != nil {
		unaryInterceptor, streamInterceptor := NewServerInterceptor(p, tp)
		streams = append(streams, streamInterceptor)
		unaries = append(unaries, unaryInterceptor)
	}
	return unaries, streams
}
