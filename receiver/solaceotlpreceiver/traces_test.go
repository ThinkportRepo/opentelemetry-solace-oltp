package solaceotlpreceiver_test

import (
	"context"
	"testing"

	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver"
	"github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver"
	"solace.dev/go/messaging/pkg/solace/resource"
)

func TestTracesReceiver_StartShutdown(t *testing.T) {
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
	}

	// Consumer and settings
	consumer := consumertest.NewNop()
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
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
	recv, err := solaceotlpreceiver.NewTracesReceiver(settings, cfg, consumer, mockMessagingService)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Start testen
	err = recv.Start(context.Background(), componenttest.NewNopHost())
	require.NoError(t, err)

	// Shutdown testen
	err = recv.Shutdown(context.Background())
	require.NoError(t, err)

	recv.QueueConsumer = mockQueueConsumer
}

func TestTracesReceiver_HandleMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Mock setup
	mockInboundMessage := mocks.NewMockInboundMessage(ctrl)

	// Test payload
	testPayload := []byte(`{
		"resourceSpans": [{
			"resource": {
				"attributes": [{
					"key": "service.name",
					"value": { "stringValue": "test-service" }
				}]
			},
			"scopeSpans": [{
				"spans": [{
					"traceId": "00000000000000000000000000000000",
					"spanId": "0000000000000000",
					"name": "test-span",
					"kind": 1
				}]
			}]
		}]
	}`)

	// Expected behavior
	mockInboundMessage.EXPECT().
		GetPayloadAsBytes().
		Return(testPayload, true)

	// Consumer and settings
	consumer := consumertest.NewNop()
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.BuildInfo{},
	}

	// Create receiver
	recv, err := solaceotlpreceiver.NewTracesReceiver(settings, &solaceotlpreceiver.Config{}, consumer)
	require.NoError(t, err)
	require.NotNil(t, recv)

	// Test message handler
	recv.HandleMessage(mockInboundMessage)
}

func TestNewTracesReceiver(t *testing.T) {
	settings := receiver.CreateSettings{
		ID:                component.NewID("solaceotlp"),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.BuildInfo{},
	}
	config := &solaceotlpreceiver.Config{
		Endpoint: "tcp://localhost:55555",
		Queue:    "otel-traces",
		Username: "default",
		Password: "default",
		VPN:      "default",
	}
	var consumer consumer.Traces

	receiver, err := solaceotlpreceiver.NewTracesReceiver(settings, config, consumer)
	if err != nil {
		t.Fatalf("NewTracesReceiver returned error: %v", err)
	}
	if receiver == nil {
		t.Fatal("NewTracesReceiver returned nil receiver")
	}
	if receiver.GetVPN() != "default" {
		t.Errorf("Expected VPN to be 'default', got %s", receiver.GetVPN())
	}
}
