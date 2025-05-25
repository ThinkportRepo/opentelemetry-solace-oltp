package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/resource"
)

type TraceData struct {
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
	ParentSpanID string                 `json:"parent_span_id,omitempty"`
	Name         string                 `json:"name"`
	Attributes   map[string]interface{} `json:"attributes"`
	Events       []EventData            `json:"events,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
}

type EventData struct {
	Name       string                 `json:"name"`
	Attributes map[string]interface{} `json:"attributes"`
	Time       time.Time              `json:"time"`
}

func loadTruststorePEMs(truststoreDir string) ([]byte, error) {
	var combinedPEM []byte
	files, err := os.ReadDir(truststoreDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read truststore directory: %v", err)
	}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pem" {
			pemPath := filepath.Join(truststoreDir, file.Name())
			pemData, err := os.ReadFile(pemPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read PEM file %s: %v", pemPath, err)
			}
			combinedPEM = append(combinedPEM, pemData...)
		}
	}
	if len(combinedPEM) == 0 {
		return nil, fmt.Errorf("no PEM files found in truststore directory")
	}
	return combinedPEM, nil
}

func initSolaceMessaging() (solace.MessagingService, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get Solace connection details
	host := os.Getenv("SOLACE_HOST")
	username := os.Getenv("SOLACE_USERNAME")
	password := os.Getenv("SOLACE_PASSWORD")
	vpnName := os.Getenv("SOLACE_VPN")

	fmt.Println("host", host)
	fmt.Println("username", username)
	fmt.Println("password", password)
	fmt.Println("vpnName", vpnName)

	if host == "" {
		host = "tcps://localhost:55443"
	}

	truststorePath := filepath.Join("truststore")

	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                   host,
		config.ServicePropertyVPNName:                       vpnName,
		config.AuthenticationPropertySchemeBasicUserName:    username,
		config.AuthenticationPropertySchemeBasicPassword:    password,
		config.TransportLayerSecurityPropertyTrustStorePath: truststorePath,
	}

	messagingService, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(brokerConfig).
		WithTransportSecurityStrategy(
			config.NewTransportSecurityStrategy().WithCertificateValidation(true, false, "", ""),
		).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to create messaging service: %v", err)
	}

	// Connect to the messaging service
	if err := messagingService.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to messaging service: %v", err)
	}

	return messagingService, nil
}

func generateTestData(ctx context.Context, messagingService solace.MessagingService) error {
	tracer := otel.Tracer("test-otlp-sender")

	// Get topic from environment
	topic := os.Getenv("SOLACE_TOPIC")
	if topic == "" {
		topic = "default/topic"
	}

	// Create root span
	ctx, rootSpan := tracer.Start(ctx, "test_operation")
	defer rootSpan.End()

	// Set attributes for root span
	rootSpan.SetAttributes(
		attribute.String("test.attribute", "root_value"),
		attribute.String("solace.topic", topic),
	)

	// Create child span
	ctx, childSpan := tracer.Start(ctx, "child_operation")

	// Set attributes for child span
	childSpan.SetAttributes(
		attribute.String("test.attribute", "child_value"),
		attribute.String("solace.topic", topic),
	)

	// Simulate an operation
	time.Sleep(100 * time.Millisecond)

	// Add event to child span
	childSpan.AddEvent("test_event", trace.WithAttributes(
		attribute.String("event.attribute", "test_value"),
		attribute.String("solace.topic", topic),
		attribute.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	))

	// Create trace data for child span
	childTraceData := TraceData{
		TraceID:      childSpan.SpanContext().TraceID().String(),
		SpanID:       childSpan.SpanContext().SpanID().String(),
		ParentSpanID: rootSpan.SpanContext().SpanID().String(),
		Name:         "child_operation",
		Attributes: map[string]interface{}{
			"test.attribute": "child_value",
			"solace.topic":   topic,
		},
		Events: []EventData{
			{
				Name: "test_event",
				Attributes: map[string]interface{}{
					"event.attribute": "test_value",
					"solace.topic":    topic,
					"timestamp":       time.Now().UTC().Format(time.RFC3339),
				},
				Time: time.Now(),
			},
		},
		StartTime: time.Now().Add(-100 * time.Millisecond),
		EndTime:   time.Now(),
	}

	// Serialize and send child span data
	childData, err := json.Marshal(childTraceData)
	if err != nil {
		return fmt.Errorf("failed to marshal child trace data: %v", err)
	}

	// Create and send message
	messageBuilder := messagingService.MessageBuilder()
	msg, err := messageBuilder.BuildWithByteArrayPayload(childData)
	if err != nil {
		return fmt.Errorf("failed to build message: %v", err)
	}

	// Create direct message publisher
	publisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		return fmt.Errorf("failed to create publisher: %v", err)
	}

	// Start the publisher
	if err := publisher.Start(); err != nil {
		return fmt.Errorf("failed to start publisher: %v", err)
	}
	defer publisher.Terminate(1 * time.Second)

	// Publish the message
	if err := publisher.Publish(msg, resource.TopicOf(topic)); err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	childSpan.End()

	// Create trace data for root span
	rootTraceData := TraceData{
		TraceID: rootSpan.SpanContext().TraceID().String(),
		SpanID:  rootSpan.SpanContext().SpanID().String(),
		Name:    "test_operation",
		Attributes: map[string]interface{}{
			"test.attribute": "root_value",
			"solace.topic":   topic,
		},
		Events: []EventData{
			{
				Name: "root_event",
				Attributes: map[string]interface{}{
					"event.attribute": "root_value",
					"solace.topic":    topic,
					"timestamp":       time.Now().UTC().Format(time.RFC3339),
				},
				Time: time.Now(),
			},
		},
		StartTime: time.Now().Add(-200 * time.Millisecond),
		EndTime:   time.Now(),
	}

	// Serialize and send root span data
	rootData, err := json.Marshal(rootTraceData)
	if err != nil {
		return fmt.Errorf("failed to marshal root trace data: %v", err)
	}

	msg, err = messageBuilder.BuildWithByteArrayPayload(rootData)
	if err != nil {
		return fmt.Errorf("failed to build root message: %v", err)
	}

	if err := publisher.Publish(msg, resource.TopicOf(topic)); err != nil {
		return fmt.Errorf("failed to publish root message: %v", err)
	}

	return nil
}

func main() {
	// Initialize Solace messaging service
	messagingService, err := initSolaceMessaging()
	if err != nil {
		log.Fatalf("Failed to initialize Solace messaging: %v", err)
	}
	defer messagingService.Disconnect()

	// Create context
	ctx := context.Background()

	// Generate test data
	log.Println("Starting test OTLP sender …")
	if err := generateTestData(ctx, messagingService); err != nil {
		log.Fatalf("Error generating test data: %v", err)
	}
	log.Println("Test data successfully generated and sent.")
}
