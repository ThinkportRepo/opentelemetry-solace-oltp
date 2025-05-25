package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	collector_trace_v1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	common_v1 "go.opentelemetry.io/proto/otlp/common/v1"
	resource_v1 "go.opentelemetry.io/proto/otlp/resource/v1"
	trace_v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	solaceresource "solace.dev/go/messaging/pkg/solace/resource"
)

func initSolaceMessaging() (messaging.MessagingService, error) {
	_ = godotenv.Load()
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

func sendTraceMessage(ms messaging.MessagingService, topic string) error {
	traceID := "0123456789abcdef0123456789abcdef"
	spanID := "abcdef0123456789"
	parentSpanID := "0000000000000000"
	traceIDBytes, _ := hex.DecodeString(traceID)
	spanIDBytes, _ := hex.DecodeString(spanID)
	parentSpanIDBytes, _ := hex.DecodeString(parentSpanID)

	span := &trace_v1.Span{
		TraceId:           traceIDBytes,
		SpanId:            spanIDBytes,
		ParentSpanId:      parentSpanIDBytes,
		Name:              "Test trace from Go to Solace receiver",
		Kind:              trace_v1.Span_SPAN_KIND_INTERNAL,
		StartTimeUnixNano: uint64(time.Now().UnixNano()),
		EndTimeUnixNano:   uint64(time.Now().Add(500 * time.Millisecond).UnixNano()),
		Attributes: []*common_v1.KeyValue{
			{Key: "custom.key", Value: &common_v1.AnyValue{Value: &common_v1.AnyValue_StringValue{StringValue: "custom-value"}}},
		},
	}

	exportRequest := &collector_trace_v1.ExportTraceServiceRequest{
		ResourceSpans: []*trace_v1.ResourceSpans{
			{
				Resource: &resource_v1.Resource{
					Attributes: []*common_v1.KeyValue{
						{Key: "service.name", Value: &common_v1.AnyValue{Value: &common_v1.AnyValue_StringValue{StringValue: "solace-trace-example"}}},
					},
				},
				ScopeSpans: []*trace_v1.ScopeSpans{{Spans: []*trace_v1.Span{span}}},
			},
		},
	}

	data, err := proto.Marshal(exportRequest)
	if err != nil {
		return fmt.Errorf("Failed to marshal OTLP trace: %v", err)
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
	topic := os.Getenv("SOLACE_TRACE_TOPIC")
	if topic == "" {
		log.Fatal("Please set SOLACE_TRACE_TOPIC")
	}
	ms, err := initSolaceMessaging()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer ms.Disconnect()

	if err := sendTraceMessage(ms, topic); err != nil {
		log.Fatalf("%v", err)
	}
	log.Println("Trace sent successfully.")
}
