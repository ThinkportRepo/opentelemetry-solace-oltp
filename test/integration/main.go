package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func initTracer() (func(), error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Configure OTLP Exporter
	ctx := context.Background()

	// Get all necessary credentials from environment variables
	endpoint := os.Getenv("SOLACE_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:4317"
	}

	username := os.Getenv("SOLACE_USERNAME")
	password := os.Getenv("SOLACE_PASSWORD")

	// VPN and TLS configuration
	useTLS := os.Getenv("USE_TLS") == "true"
	caCertPath := os.Getenv("CA_CERT_PATH")
	vpnHostname := os.Getenv("VPN_HOSTNAME")

	var dialOption grpc.DialOption
	if useTLS {
		// Load TLS configuration for VPN
		tlsConfig := &tls.Config{
			ServerName: vpnHostname,
		}

		if caCertPath != "" {
			// Load CA certificate if provided
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %v", err)
			}
			tlsConfig.RootCAs = x509.NewCertPool()
			if !tlsConfig.RootCAs.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to append CA certificate")
			}
		}

		dialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		dialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	// Configure client with credentials and VPN settings
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(dialOption),
		otlptracegrpc.WithHeaders(map[string]string{
			"username": username,
			"password": password,
		}),
		otlptracegrpc.WithTimeout(30*time.Second),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %v", err)
	}

	// Create Resource with service information from environment variables
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "test-otlp-sender"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %v", err)
	}

	// Configure Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// Cleanup function
	cleanup := func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}

	return cleanup, nil
}

func generateTestData(ctx context.Context) error {
	tracer := otel.Tracer("test-otlp-sender")

	// Create root span
	ctx, rootSpan := tracer.Start(ctx, "test_operation")
	defer rootSpan.End()

	// Set attributes for root span
	rootSpan.SetAttributes(attribute.String("test.attribute", "root_value"))

	// Create child span
	_, childSpan := tracer.Start(ctx, "child_operation")

	// Set attributes for child span
	childSpan.SetAttributes(attribute.String("test.attribute", "child_value"))

	// Simulate an operation
	time.Sleep(100 * time.Millisecond)

	// Add event to child span
	childSpan.AddEvent("test_event", trace.WithAttributes(
		attribute.String("event.attribute", "test_value"),
		attribute.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	))

	childSpan.End()

	// Add event to root span
	rootSpan.AddEvent("root_event", trace.WithAttributes(
		attribute.String("event.attribute", "root_value"),
		attribute.String("timestamp", time.Now().UTC().Format(time.RFC3339)),
	))

	return nil
}

func main() {
	// Initialize tracer
	cleanup, err := initTracer()
	if err != nil {
		log.Fatalf("Tracer initialization failed: %v", err)
	}
	defer cleanup()

	// Create context
	ctx := context.Background()

	// Generate test data
	log.Println("Starting test OTLP sender...")
	if err := generateTestData(ctx); err != nil {
		log.Fatalf("Error generating test data: %v", err)
	}
	log.Println("Test data successfully generated and sent.")
}
