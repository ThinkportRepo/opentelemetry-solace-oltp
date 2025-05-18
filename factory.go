package solacetracereceiver

import (
	"github.com/open-telemetry/opentelemetry-collector/receiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		"solacetrace",
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver),
	)
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID("solacetrace")),
		// Setze hier Standardwerte für die Konfiguration
	}
}

func createTracesReceiver(
	ctx context.Context,
	settings receiver.CreateSettings,
	cfg config.Receiver,
	nextConsumer consumer.Traces,
) (receiver.Traces, error) {
	return newSolaceTraceReceiver(cfg.(*Config), nextConsumer)
}
