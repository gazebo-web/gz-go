# Telemetry

## Tracing

You can use this package to enable tracing in your project. The current implementation was designed using
[OpenTelemetry](https://opentelemetry.io/) and [Jaeger](https://www.jaegertracing.io/).

## Quickstart

### Configuration

```go
// Embed telemetry.TracingConfig to your configuration struct
type MyAwesomeProject struct {
    Tracing telemetry.TracingConfig
}

// Or use telemetry.ParseTracingConfig()
cfg, err := telemetry.ParseTracingConfig()
if err != nil {
    // Handle error
}
```

### Initialize tracing components

```go
// Using the config parsed in the previous step:
propagator, tracerProvider, err := telemetry.InitializeTracing(cfg)
if err != nil {
    // Handle error
}
```

### HTTP

In order to use tracing for your HTTP endpoints, you can wrap the given endpoint with tracing:

```go
endpoint = telemetry.WrapHandlerHTTP(endpoint, "http.CreateUser", propagator, tracerProvider)
```

NOTE: As it is right now, each endpoint needs to be wrapped independently. Support for HTTP routers will be added
in future iterations.

### gRPC

#### Client

```go
// Initialize interceptors to add to your gRPC client:
unaryInterceptor, streamInterceptor := telemetry.NewClientInterceptor(propagator, tracerProvider)
```

#### Server

```go
// We can also add interceptors to our gRPC server:
unaryInterceptor, streamInterceptor := telemetry.NewServerInterceptor(propagator, tracerProvider)
```

This package also includes some helper functions to make you tracing initialization easier when adding middlewares to
your gRPC servers and clients.