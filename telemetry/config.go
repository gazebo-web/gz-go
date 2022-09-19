package telemetry

import (
	"errors"
	"github.com/caarlos0/env/v6"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig defines configuration values to customize how Tracing is initialized for different services.
type TracingConfig struct {
	// Service contains the service name that will be used when generating traces. It's usually the name of the service
	// that's using this library.
	Service string `env:"SERVICE,notEmpty"`

	// Environment defines the environment being traced. Defaults to staging.
	Environment string `env:"ENVIRONMENT" envDefault:"development"`

	// Enabled defines if tracing should be enabled.
	Enabled bool `env:"ENABLED" envDefault:"false"`

	// ExportingStrategy contains the name of the strategy used to export traces. Defaults to collector.
	// Possible values: collector, agent.
	ExportingStrategy string `env:"EXPORTING_STRATEGY" envDefault:"collector"`

	// CollectorURL defines the URL traces should be sent to. If Enabled is true, this value
	// must be set.
	CollectorURL string `env:"COLLECTOR_URL" envDefault:"http://localhost:14268/api/traces"`

	// AgentHost defines the address this service should send traces to. If Enabled is true, this value
	// must be set.
	AgentHost string `env:"AGENT_HOST" envDefault:"localhost"`

	// AgentPort defines the port used alongside AgentHost. If Enabled is true, this value must be set.
	AgentPort string `env:"AGENT_PORT" envDefault:"6831"`
}

// ParseTracingConfig parses TracingConfig from environment variables.
// All the environment variables specified in TracingConfig are prepend with the TRACING_ prefix.
func ParseTracingConfig() (TracingConfig, error) {
	var cfg TracingConfig
	if err := env.Parse(&cfg, env.Options{
		Prefix: "TRACING_",
	}); err != nil {
		return TracingConfig{}, err
	}
	return cfg, nil
}

// InitializeTracing initializes the components used for exporting traces in a project using the config defined by TracingConfig.
// If TracingConfig.Enabled is set to false, it returns nil values.
func InitializeTracing(cfg TracingConfig) (propagation.TextMapPropagator, trace.TracerProvider, error) {
	if !cfg.Enabled {
		return nil, nil, nil
	}

	var propagator propagation.TextMapPropagator
	var tracerProvider trace.TracerProvider
	var err error
	switch cfg.ExportingStrategy {
	case "collector":
		tracerProvider, err = NewJaegerTracerProviderCollector(cfg.Service, cfg.CollectorURL, cfg.Environment)
	case "agent":
		tracerProvider, err = NewJaegerTracerProviderAgent(cfg.Service, cfg.AgentHost, cfg.AgentPort, cfg.Environment)
	default:
		return nil, nil, errors.New("invalid exporting strategy")
	}

	if err != nil {
		return nil, nil, err
	}

	return propagator, tracerProvider, nil
}
