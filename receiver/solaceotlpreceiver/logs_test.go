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
	"solace.dev/go/messaging/pkg/solace/resource"
)

// nopHost is an empty host mock for tests
// Implements the component.Host interface
type nopHost struct{}

func (nopHost) ReportFatalError(error)                                                    {}
func (nopHost) GetFactory(_ component.Kind, _ component.Type) component.Factory           { return nil }
func (nopHost) GetExtensions() map[component.ID]component.Component                       { return nil }
func (nopHost) GetExporters() map[component.DataType]map[component.ID]component.Component { return nil }

func TestLogsReceiver_StartShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock setup
	mockMessagingService := mocks.NewMockMessagingService(ctrl)
	mockQueueConsumer := mocks.NewMockQueueConsumer(ctrl)
	mockQueueConsumerBuilder := mocks.NewMockQueueConsumerBuilder(ctrl)

	// Configuration
	cfg := &solaceotlpreceiver.Config{
		Endpoint: "tcp://localhost:55555",
		Queue:    "test-queue",
		Username: "user",
		Password: "pass",
		VPN:      "default",
	}

	// Consumer and settings
	// Consumer mock (can be nil as it's not used in the test)
	var consumer consumer.Logs
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
		BuildInfo:         component.BuildInfo{},
	}

	// Expected behavior
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

	// Create receiver
	recv, err := solaceotlpreceiver.NewLogsReceiver(settings, cfg, consumer, mockMessagingService)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Test start
	err = recv.Start(context.Background(), nopHost{})
	require.NoError(t, err)

	recv.QueueConsumer = mockQueueConsumer
	// Test shutdown
	err = recv.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestLogsReceiver_HandleMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock setup
	mockInboundMessage := mocks.NewMockInboundMessage(ctrl)

	// Test payload
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

	// Expected behavior
	mockInboundMessage.EXPECT().
		GetPayloadAsBytes().
		Return(testPayload, true)

	// Consumer and settings
	var consumer consumer.Logs
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: component.TelemetrySettings{Logger: zap.NewNop()},
		BuildInfo:         component.BuildInfo{},
	}

	// Create receiver
	recv, err := solaceotlpreceiver.NewLogsReceiver(settings, &solaceotlpreceiver.Config{}, consumer)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Test message handler
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
