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
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	solaceresource "solace.dev/go/messaging/pkg/solace/resource"
)

type solaceExporter struct {
	messagingService solace.MessagingService
	topic            string
}

func newSolaceExporter(messagingService solace.MessagingService, topic string) *solaceExporter {
	return &solaceExporter{
		messagingService: messagingService,
		topic:            topic,
	}
}

func (e *solaceExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	for _, span := range spans {
		// Create message with span data
		messageBuilder := e.messagingService.MessageBuilder()
		msg, err := messageBuilder.BuildWithStringPayload(fmt.Sprintf(
			`{"trace_id":"%s","span_id":"%s","parent_span_id":"%s","name":"%s","kind":%d,"start_time":%d,"end_time":%d,"status":{"code":%d,"message":"%s"}}`,
			span.SpanContext().TraceID().String(),
			span.SpanContext().SpanID().String(),
			span.Parent().SpanID().String(),
			span.Name(),
			span.SpanKind(),
			span.StartTime().UnixNano(),
			span.EndTime().UnixNano(),
			span.Status().Code,
			span.Status().Description,
		))
		if err != nil {
			return fmt.Errorf("failed to build message: %v", err)
		}

		// Create publisher
		publisher, err := e.messagingService.CreateDirectMessagePublisherBuilder().Build()
		if err != nil {
			return fmt.Errorf("failed to create publisher: %v", err)
		}

		// Start publisher
		if err := publisher.Start(); err != nil {
			return fmt.Errorf("failed to start publisher: %v", err)
		}
		defer publisher.Terminate(1 * time.Second)

		// Publish message
		if err := publisher.Publish(msg, solaceresource.TopicOf(e.topic)); err != nil {
			return fmt.Errorf("failed to publish message: %v", err)
		}
	}
	return nil
}

func (e *solaceExporter) Shutdown(ctx context.Context) error {
	return nil
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

// SeverityNumber repräsentiert die OTLP Log Severity Levels
type SeverityNumber int32

const (
	SEVERITY_NUMBER_UNSPECIFIED SeverityNumber = 0
	SEVERITY_NUMBER_TRACE       SeverityNumber = 1
	SEVERITY_NUMBER_DEBUG       SeverityNumber = 5
	SEVERITY_NUMBER_INFO        SeverityNumber = 9
	SEVERITY_NUMBER_WARN        SeverityNumber = 13
	SEVERITY_NUMBER_ERROR       SeverityNumber = 17
	SEVERITY_NUMBER_FATAL       SeverityNumber = 21
)

// KeyValue repräsentiert ein OTLP KeyValue Attribut
type KeyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// LogRecord repräsentiert einen OTLP Log Record
type LogRecord struct {
	TimeUnixNano         int64      `json:"time_unix_nano"`
	ObservedTimeUnixNano int64      `json:"observed_time_unix_nano"`
	SeverityNumber       int32      `json:"severity_number"`
	SeverityText         string     `json:"severity_text"`
	Body                 string     `json:"body"`
	Attributes           []KeyValue `json:"attributes"`
	TraceID              string     `json:"trace_id"`
	SpanID               string     `json:"span_id"`
	EventName            string     `json:"event_name,omitempty"`
}

func getSeverityNumber(severity string) SeverityNumber {
	switch severity {
	case "TRACE":
		return SEVERITY_NUMBER_TRACE
	case "DEBUG":
		return SEVERITY_NUMBER_DEBUG
	case "INFO":
		return SEVERITY_NUMBER_INFO
	case "WARN":
		return SEVERITY_NUMBER_WARN
	case "ERROR":
		return SEVERITY_NUMBER_ERROR
	case "FATAL":
		return SEVERITY_NUMBER_FATAL
	default:
		return SEVERITY_NUMBER_UNSPECIFIED
	}
}

func sendLogMessage(messagingService solace.MessagingService, topic string, traceID, spanID string, severity, message string, attributes map[string]string) error {
	// Create message with log data
	messageBuilder := messagingService.MessageBuilder()

	// Convert attributes to KeyValue array
	keyValues := make([]KeyValue, 0, len(attributes))
	for k, v := range attributes {
		keyValues = append(keyValues, KeyValue{
			Key:   k,
			Value: v,
		})
	}

	// Create log record
	now := time.Now()
	logRecord := LogRecord{
		TimeUnixNano:         now.UnixNano(),
		ObservedTimeUnixNano: now.UnixNano(),
		SeverityNumber:       int32(getSeverityNumber(severity)),
		SeverityText:         severity,
		Body:                 message,
		Attributes:           keyValues,
		TraceID:              traceID,
		SpanID:               spanID,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(logRecord)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %v", err)
	}

	msg, err := messageBuilder.BuildWithStringPayload(string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to build message: %v", err)
	}

	// Create publisher
	publisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		return fmt.Errorf("failed to create publisher: %v", err)
	}

	// Start publisher
	if err := publisher.Start(); err != nil {
		return fmt.Errorf("failed to start publisher: %v", err)
	}
	defer publisher.Terminate(1 * time.Second)

	// Publish message
	if err := publisher.Publish(msg, solaceresource.TopicOf(topic)); err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
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

	// Set attributes for root span
	rootSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.url", "/api/v1/orders"),
		attribute.String("http.user_agent", "Mozilla/5.0"),
		attribute.String("http.request_id", "req-123456"),
		attribute.String("solace.topic", topic),
	)

	// Send log message for root span
	if err := sendLogMessage(
		messagingService,
		topic,
		rootSpan.SpanContext().TraceID().String(),
		rootSpan.SpanContext().SpanID().String(),
		"INFO",
		"Processing new API request",
		map[string]string{
			"http.method": "POST",
			"http.url":    "/api/v1/orders",
		},
	); err != nil {
		return fmt.Errorf("failed to send log message: %v", err)
	}

	// Simulate authentication
	ctx, authSpan := tracer.Start(ctx, "authenticate_user")
	authSpan.SetAttributes(
		attribute.String("auth.method", "jwt"),
		attribute.String("auth.user_id", "user-789"),
	)

	// Send error log for authentication
	if err := sendLogMessage(
		messagingService,
		topic,
		authSpan.SpanContext().TraceID().String(),
		authSpan.SpanContext().SpanID().String(),
		"ERROR",
		"User not found: id=42",
		map[string]string{
			"user.id":     "42",
			"error.code":  "USER_NOT_FOUND",
			"auth.method": "jwt",
		},
	); err != nil {
		return fmt.Errorf("failed to send log message: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	authSpan.End()

	// Simulate database operation
	ctx, dbSpan := tracer.Start(ctx, "database_operation")
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.name", "orders_db"),
		attribute.String("db.operation", "insert"),
	)
	time.Sleep(100 * time.Millisecond)
	dbSpan.End()

	// Simulate external service call
	_, serviceSpan := tracer.Start(ctx, "external_service_call")
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

	// Force flush to ensure all spans are exported
	if tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider); ok {
		if err := tp.ForceFlush(ctx); err != nil {
			return fmt.Errorf("failed to flush traces: %v", err)
		}
	}

	return nil
}

func initTracer(messagingService solace.MessagingService) error {
	// Get topic from environment
	topic := os.Getenv("SOLACE_TOPIC")
	if topic == "" {
		topic = "default/topic"
	}

	// Create Solace exporter
	exporter := newSolaceExporter(messagingService, topic)

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("test-otlp-sender"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %v", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return nil
}

func main() {
	// Initialize Solace messaging
	messagingService, err := initSolaceMessaging()
	if err != nil {
		log.Fatalf("Failed to initialize Solace messaging: %v", err)
	}
	defer messagingService.Disconnect()

	// Initialize tracer
	if err := initTracer(messagingService); err != nil {
		log.Fatalf("Failed to initialize tracer: %v", err)
	}

	// Generate and send test data
	ctx := context.Background()
	if err := generateTestData(ctx, messagingService); err != nil {
		log.Fatalf("Failed to generate test data: %v", err)
	}

	log.Println("Test data sent successfully")
}
