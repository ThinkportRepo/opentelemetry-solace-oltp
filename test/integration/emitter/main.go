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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	solaceresource "solace.dev/go/messaging/pkg/solace/resource"
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

	// Create root span for API request
	ctx, rootSpan := tracer.Start(ctx, "api_request")
	defer rootSpan.End()

	// Get root span context
	rootSpanContext := rootSpan.SpanContext()
	rootTraceID := rootSpanContext.TraceID().String()
	rootSpanID := rootSpanContext.SpanID().String()

	// Set attributes for root span
	rootSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", "/api/v1/orders"),
		attribute.String("http.user_agent", "Mozilla/5.0"),
		attribute.String("http.request_id", "req-123456"),
		attribute.String("solace.topic", topic),
	)

	// Simulate authentication
	ctx, authSpan := tracer.Start(ctx, "authenticate_user")
	authSpanContext := authSpan.SpanContext()
	authSpanID := authSpanContext.SpanID().String()
	authSpan.SetAttributes(
		attribute.String("auth.method", "jwt"),
		attribute.String("auth.user_id", "user-789"),
	)
	time.Sleep(50 * time.Millisecond)
	authSpan.End()

	// Simulate database operation
	ctx, dbSpan := tracer.Start(ctx, "database_operation")
	dbSpanContext := dbSpan.SpanContext()
	dbSpanID := dbSpanContext.SpanID().String()
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "orders_db"),
		attribute.String("db.operation", "insert"),
	)
	time.Sleep(100 * time.Millisecond)
	dbSpan.End()

	// Simulate external service call
	_, serviceSpan := tracer.Start(ctx, "external_service_call")
	serviceSpanContext := serviceSpan.SpanContext()
	serviceSpanID := serviceSpanContext.SpanID().String()
	serviceSpan.SetAttributes(
		attribute.String("service.name", "payment-service"),
		attribute.String("service.operation", "process_payment"),
		attribute.String("service.version", "1.0.0"),
	)

	// Add events to service span
	serviceSpan.AddEvent("payment_processing_started", trace.WithAttributes(
		attribute.String("payment.id", "pay-123"),
		attribute.Float64("payment.amount", 99.99),
		attribute.String("currency", "EUR"),
	))

	time.Sleep(150 * time.Millisecond)

	serviceSpan.AddEvent("payment_processing_completed", trace.WithAttributes(
		attribute.String("payment.status", "success"),
		attribute.String("transaction.id", "tx-456"),
	))

	serviceSpan.End()

	// Create trace data for root span with all child spans
	rootTraceData := TraceData{
		TraceID: rootTraceID,
		SpanID:  rootSpanID,
		Name:    "api_request",
		Attributes: map[string]interface{}{
			"http.method":     "POST",
			"http.url":        "/api/v1/orders",
			"http.user_agent": "Mozilla/5.0",
			"http.request_id": "req-123456",
			"solace.topic":    topic,
		},
		Events: []EventData{
			{
				Name: "request_received",
				Attributes: map[string]interface{}{
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				},
				Time: time.Now(),
			},
		},
		StartTime: time.Now().Add(-300 * time.Millisecond),
		EndTime:   time.Now(),
	}

	// Create trace data for child spans
	authTraceData := TraceData{
		TraceID:      rootTraceID,
		SpanID:       authSpanID,
		ParentSpanID: rootSpanID,
		Name:         "authenticate_user",
		Attributes: map[string]interface{}{
			"auth.method":  "jwt",
			"auth.user_id": "user-789",
		},
		StartTime: time.Now().Add(-250 * time.Millisecond),
		EndTime:   time.Now().Add(-200 * time.Millisecond),
	}

	dbTraceData := TraceData{
		TraceID:      rootTraceID,
		SpanID:       dbSpanID,
		ParentSpanID: rootSpanID,
		Name:         "database_operation",
		Attributes: map[string]interface{}{
			"db.system":    "postgresql",
			"db.name":      "orders_db",
			"db.operation": "insert",
		},
		StartTime: time.Now().Add(-200 * time.Millisecond),
		EndTime:   time.Now().Add(-100 * time.Millisecond),
	}

	serviceTraceData := TraceData{
		TraceID:      rootTraceID,
		SpanID:       serviceSpanID,
		ParentSpanID: rootSpanID,
		Name:         "external_service_call",
		Attributes: map[string]interface{}{
			"service.name":      "payment-service",
			"service.operation": "process_payment",
			"service.version":   "1.0.0",
		},
		Events: []EventData{
			{
				Name: "payment_processing_started",
				Attributes: map[string]interface{}{
					"payment.id":     "pay-123",
					"payment.amount": 99.99,
					"currency":       "EUR",
				},
				Time: time.Now().Add(-150 * time.Millisecond),
			},
			{
				Name: "payment_processing_completed",
				Attributes: map[string]interface{}{
					"payment.status": "success",
					"transaction.id": "tx-456",
				},
				Time: time.Now(),
			},
		},
		StartTime: time.Now().Add(-150 * time.Millisecond),
		EndTime:   time.Now(),
	}

	// Serialize and send all trace data
	traceDataList := []TraceData{rootTraceData, authTraceData, dbTraceData, serviceTraceData}

	for _, traceData := range traceDataList {
		data, err := json.Marshal(traceData)
		if err != nil {
			return fmt.Errorf("failed to marshal trace data: %v", err)
		}

		// Debug output
		fmt.Printf("Sending trace: %s\n", string(data))

		messageBuilder := messagingService.MessageBuilder()
		msg, err := messageBuilder.BuildWithByteArrayPayload(data)
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
		if err := publisher.Publish(msg, solaceresource.TopicOf(topic)); err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
	}

	return nil
}

func initTracer() error {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	return nil
}

func main() {
	// Initialize tracer
	if err := initTracer(); err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}

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
