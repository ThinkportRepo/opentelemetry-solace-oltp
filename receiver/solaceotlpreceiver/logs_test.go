package solaceotlpreceiver_test

import (
	"context"
	"testing"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// nopHost ist ein leerer Host-Mock für die Tests
// Erfüllt das component.Host Interface
type nopHost struct{}

func (nopHost) ReportFatalError(error)                                                    {}
func (nopHost) GetFactory(_ component.Kind, _ component.Type) component.Factory           { return nil }
func (nopHost) GetExtensions() map[component.ID]component.Component                       { return nil }
func (nopHost) GetExporters() map[component.DataType]map[component.ID]component.Component { return nil }

func TestLogsReceiver_StartShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock-Setup
	mockMessagingService := mocks.NewMockMessagingService(ctrl)
	mockQueueConsumer := mocks.NewMockQueueConsumer(ctrl)
	mockQueueConsumerBuilder := mocks.NewMockQueueConsumerBuilder(ctrl)

	// Konfiguration
	cfg := &solaceotlpreceiver.Config{
		Endpoint: "tcp://localhost:55555",
		Queue:    "test-queue",
		Username: "user",
		Password: "pass",
		VPN:      "default",
	}

	// Consumer und Settings
	// Consumer-Mock (kann nil sein, da im Test nicht verwendet)
	var consumer consumer.Logs
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
		BuildInfo:         component.BuildInfo{},
	}

	// Erwartetes Verhalten
	mockMessagingService.EXPECT().
		Connect().
		Return(nil)

	mockMessagingService.EXPECT().
		CreateQueueConsumerBuilder().
		Return(mockQueueConsumerBuilder)

	mockQueueConsumerBuilder.EXPECT().
		WithMessageAutoAcknowledgement().
		Return(mockQueueConsumerBuilder)

	mockQueueConsumerBuilder.EXPECT().
		WithMessageListener(gomock.Any()).
		Return(mockQueueConsumerBuilder)

	mockQueueConsumerBuilder.EXPECT().
		Build(*resource.QueueDurableExclusive(cfg.Queue)).
		Return(mockQueueConsumer, nil)

	mockQueueConsumer.EXPECT().
		Start().
		Return(nil)

	mockQueueConsumer.EXPECT().
		Terminate(uint(10)).
		Return(nil)

	mockMessagingService.EXPECT().
		Disconnect().
		Return(nil)

	// Receiver erstellen
	recv, err := solaceotlpreceiver.NewLogsReceiver(settings, cfg, consumer, mockMessagingService)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Start testen
	err = recv.Start(context.Background(), nopHost{})
	require.NoError(t, err)

	recv.QueueConsumer = mockQueueConsumer
	// Shutdown testen
	err = recv.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestLogsReceiver_HandleMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock-Setup
	mockInboundMessage := mocks.NewMockInboundMessage(ctrl)

	// Test-Payload
	testPayload := []byte(`{
		"resourceLogs": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "test-service" }
				}]
			},
			"scopeLogs": [{
				"logRecords": [{
					"timeUnixNano": "1640995200000000000",
					"severityNumber": 9,
					"severityText": "INFO",
					"body": { "stringValue": "test log message" }
				}]
			}]
		}]
	}`)

	// Erwartetes Verhalten
	mockInboundMessage.EXPECT().
		GetPayloadAsBytes().
		Return(testPayload, true)

	mockInboundMessage.EXPECT().
		GetCacheRequestID().
		Return(message.CacheRequestID(0), false).AnyTimes()

	mockInboundMessage.EXPECT().
		GetCacheStatus().
		Return(message.CacheStatus(0)).AnyTimes()

	// Consumer und Settings
	var consumer consumer.Logs
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
		BuildInfo:         component.BuildInfo{},
	}

	// Receiver erstellen
	recv, err := solaceotlpreceiver.NewLogsReceiver(settings, &solaceotlpreceiver.Config{}, consumer)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Message-Handler testen
	recv.HandleMessage(mockInboundMessage)
}

func TestNewLogsReceiver(t *testing.T) {
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
		BuildInfo:         component.BuildInfo{},
	}
	config := &solaceotlpreceiver.Config{
		Endpoint: "tcp://localhost:55555",
		Queue:    "otel-logs",
		Username: "default",
		Password: "default",
		VPN:      "default",
	}
	var consumer consumer.Logs

	receiver, err := solaceotlpreceiver.NewLogsReceiver(settings, config, consumer)
	if err != nil {
		t.Fatalf("NewLogsReceiver returned error: %v", err)
	}
	if receiver == nil {
		t.Fatal("NewLogsReceiver returned nil receiver")
	}
	if receiver.GetVPN() != "default" {
		t.Errorf("Expected VPN to be 'default', got %s", receiver.GetVPN())
	}
}
