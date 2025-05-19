package solaceotlpreceiver

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr = "solaceotlp"
)

// NewFactory creates a factory for Solace OTLP receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, component.StabilityLevelAlpha),
		receiver.WithLogs(createLogsReceiver, component.StabilityLevelAlpha),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Queue: "telemetry-queue",
	}
}

func createTracesReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Traces,
) (receiver.Traces, error) {
	if consumer == nil {
		return nil, fmt.Errorf("nil consumer")
	}

	config := cfg.(*Config)
	return newTracesReceiver(params, config, consumer)
}

func createLogsReceiver(
	_ context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	if consumer == nil {
		return nil, fmt.Errorf("nil consumer")
	}

	config := cfg.(*Config)
	return newLogsReceiver(params, config, consumer)
}
