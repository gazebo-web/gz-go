package telemetry

import (
	"github.com/caarlos0/env/v6"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig defines configuration values to customize how Tracing is initialized for different services.
// All the environment variables specified in this structure are prepend with the TRACING_ prefix.
type TracingConfig struct {
	// Service contains the service name that will be used when generating traces. It's usually the name of the service
	// that's using this library.
	Service string `env:"SERVICE,notEmpty"`

	// Environment defines the environment where tracing is being performed is running on. Defaults to staging.
	Environment string `env:"ENVIRONMENT" envDefault:"staging"`

	// Enabled is set to true when tracing should be enabled for the current service.
	Enabled bool `env:"ENABLED" envDefault:"false"`

	// ExportingStrategy contains the name of the strategy that is being used for exporting tracing. Defaults to collector.
	// Available values: collector, agent.
	ExportingStrategy string `env:"EXPORTING_STRATEGY" envDefault:"collector"`

	// CollectorURL defines the URL where traces should be sent to. If Enabled is true, this value
	// must be set.
	CollectorURL string `env:"COLLECTOR_URL" envDefault:"http://localhost:14268/api/traces"`

	// AgentHost defines the address where this service should send traces to. If Enabled is true, this value
	// must be set.
	AgentHost string `env:"AGENT_HOST" envDefault:"localhost"`

	// AgentPort defines the port used alongside AgentHost. If Enabled is true, this value must be set.
	AgentPort string `env:"AGENT_PORT" envDefault:"6831"`
}

// ParseTracingConfig parses TracingConfig from environment variables.
func ParseTracingConfig() (TracingConfig, error) {
	var cfg TracingConfig
	if err := env.Parse(&cfg, env.Options{
		Prefix: "TRACING_",
	}); err != nil {
		return TracingConfig{}, err
	}
	return cfg, nil
}

// InitializeTracing initializes tracing defined by TracingConfig. If TracingConfig.Enabled is set to false, it returns
// nil values.
func InitializeTracing(cfg TracingConfig) (propagation.TextMapPropagator, trace.TracerProvider, error) {
	if !cfg.Enabled {
		return nil, nil, nil
	}

	var propagator propagation.TextMapPropagator
	var tracerProvider trace.TracerProvider
	var err error
	switch cfg.ExportingStrategy {
	case "collector":
		tracerProvider, err = NewJaegerTracerProviderAgent(cfg.Service, cfg.AgentHost, cfg.AgentPort, cfg.Environment)
	case "agent":
		tracerProvider, err = NewJaegerTracerProviderCollector(cfg.Service, cfg.CollectorURL, cfg.Environment)
	}

	if err != nil {
		return nil, nil, err
	}

	return propagator, tracerProvider, nil
}
