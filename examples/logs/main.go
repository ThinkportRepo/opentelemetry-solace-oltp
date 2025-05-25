package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	collector_logs_v1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	common_v1 "go.opentelemetry.io/proto/otlp/common/v1"
	logs_v1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resource_v1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	solaceresource "solace.dev/go/messaging/pkg/solace/resource"
)

func initSolaceMessaging() (messaging.MessagingService, error) {
	_ = godotenv.Load()
	// Read connection details from environment
	host := os.Getenv("SOLACE_HOST")
	vpn := os.Getenv("SOLACE_VPN")
	username := os.Getenv("SOLACE_USERNAME")
	password := os.Getenv("SOLACE_PASSWORD")
	if host == "" || vpn == "" || username == "" || password == "" {
		return nil, fmt.Errorf("Please set SOLACE_HOST, SOLACE_VPN, SOLACE_USERNAME, SOLACE_PASSWORD")
	}
	truststorePath := "truststore"
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                   host,
		config.ServicePropertyVPNName:                       vpn,
		config.AuthenticationPropertySchemeBasicUserName:    username,
		config.AuthenticationPropertySchemeBasicPassword:    password,
		config.TransportLayerSecurityPropertyTrustStorePath: truststorePath,
	}
	ms, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(brokerConfig).
		WithTransportSecurityStrategy(
			config.NewTransportSecurityStrategy().WithCertificateValidation(true, false, "", ""),
		).
		Build()
	if err != nil {
		return nil, fmt.Errorf("Failed to create messaging service: %v", err)
	}
	if err := ms.Connect(); err != nil {
		return nil, fmt.Errorf("Failed to connect to Solace: %v", err)
	}
	return ms, nil
}

func sendLogMessage(ms messaging.MessagingService, topic string) error {
	traceID := "0123456789abcdef0123456789abcdef"
	spanID := "abcdef0123456789"
	traceIDBytes, _ := hex.DecodeString(traceID)
	spanIDBytes, _ := hex.DecodeString(spanID)

	logRecord := &logs_v1.LogRecord{
		TimeUnixNano:   uint64(time.Now().UnixNano()),
		SeverityNumber: logs_v1.SeverityNumber_INFO,
		SeverityText:   "INFO",
		Body:           &common_v1.AnyValue{Value: &common_v1.AnyValue_StringValue{StringValue: "Test log from Go to Solace receiver"}},
		TraceId:        traceIDBytes,
		SpanId:         spanIDBytes,
		Attributes: []*common_v1.KeyValue{
			{Key: "custom.key", Value: &common_v1.AnyValue{Value: &common_v1.AnyValue_StringValue{StringValue: "custom-value"}}},
		},
	}

	exportRequest := &collector_logs_v1.ExportLogsServiceRequest{
		ResourceLogs: []*logs_v1.ResourceLogs{
			{
				Resource: &resource_v1.Resource{
					Attributes: []*common_v1.KeyValue{
						{Key: "service.name", Value: &common_v1.AnyValue{Value: &common_v1.AnyValue_StringValue{StringValue: "solace-log-example"}}},
					},
				},
				ScopeLogs: []*logs_v1.ScopeLogs{{LogRecords: []*logs_v1.LogRecord{logRecord}}},
			},
		},
	}

	data, err := proto.Marshal(exportRequest)
	if err != nil {
		return fmt.Errorf("Failed to marshal OTLP log: %v", err)
	}

	msg, err := ms.MessageBuilder().BuildWithStringPayload(base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return fmt.Errorf("Failed to build message: %v", err)
	}

	publisher, err := ms.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		return fmt.Errorf("Failed to create publisher: %v", err)
	}
	if err := publisher.Start(); err != nil {
		return fmt.Errorf("Failed to start publisher: %v", err)
	}
	defer publisher.Terminate(1 * time.Second)

	if err := publisher.Publish(msg, solaceresource.TopicOf(topic)); err != nil {
		return fmt.Errorf("Failed to publish message: %v", err)
	}

	return nil
}

func main() {
	topic := os.Getenv("SOLACE_LOG_TOPIC")
	if topic == "" {
		log.Fatal("Please set SOLACE_LOG_TOPIC")
	}
	ms, err := initSolaceMessaging()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer ms.Disconnect()

	if err := sendLogMessage(ms, topic); err != nil {
		log.Fatalf("%v", err)
	}
	log.Println("Log sent successfully.")
}
