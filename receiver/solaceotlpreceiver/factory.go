package solaceotlpreceiver

import (
	"context"
	"fmt"

	solaceconfig "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/config"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var (
	typeStr = component.MustNewType("solaceotlp")
)

// NewFactory creates a factory for Solace OTLP receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelStable),
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelAlpha),
	)
}

// createDefaultConfig creates the default configuration for the receiver
func createDefaultConfig() component.Config {
	return &solaceconfig.Config{
		Queue: "telemetry",
	}
}

// createTracesReceiver creates a new traces receiver
func createTracesReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	if consumer == nil {
		return nil, fmt.Errorf("nil consumer")
	}

	conf := cfg.(*solaceconfig.Config)
	receiver, err := NewReceiver(settings, conf, nil, consumer)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}

// createLogsReceiver creates a new logs receiver
func createLogsReceiver(
	_ context.Context,
	settings receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	if consumer == nil {
		return nil, fmt.Errorf("nil consumer")
	}

	conf := cfg.(*solaceconfig.Config)
	receiver, err := NewReceiver(settings, conf, consumer, nil)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}
