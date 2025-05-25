package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/resource"
)

func main() {
	godotenv.Load(".env")

	endpoint := os.Getenv("SOLACE_HOST")
	if endpoint == "" {
		endpoint = os.Getenv("SOLACE_HOST")
	}
	username := os.Getenv("SOLACE_USERNAME")
	password := os.Getenv("SOLACE_PASSWORD")
	vpn := os.Getenv("SOLACE_VPN")
	queue := os.Getenv("SOLACE_QUEUE")
	trustStore := os.Getenv("SOLACE_TRUST_STORE_PATH")
	if trustStore == "" {
		trustStore = "truststore"
	}

	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("VPN: %s\n", vpn)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("TrustStore: %s\n", trustStore)
	fmt.Printf("Queue: %s\n", queue)

	fmt.Println("Connecting to Solace …")
	service, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(config.ServicePropertyMap{
			config.TransportLayerPropertyHost:                   endpoint,
			config.ServicePropertyVPNName:                       vpn,
			config.AuthenticationPropertySchemeBasicUserName:    username,
			config.AuthenticationPropertySchemeBasicPassword:    password,
			config.TransportLayerSecurityPropertyTrustStorePath: trustStore,
		}).
		WithTransportSecurityStrategy(
			config.NewTransportSecurityStrategy().WithCertificateValidation(true, false, "", ""),
		).
		Build()
	if err != nil {
		log.Fatalf("Failed to build messaging service: %v", err)
	}

	if err := service.Connect(); err != nil {
		log.Fatalf("Failed to connect to Solace: %v", err)
	}
	fmt.Println("Connected.")

	consumerBuilder := service.CreatePersistentMessageReceiverBuilder()
	receiver, err := consumerBuilder.Build(resource.QueueDurableExclusive(queue))
	if err != nil {
		log.Fatalf("Failed to build queue consumer: %v", err)
	}

	if err := receiver.Start(); err != nil {
		log.Fatalf("Failed to start queue consumer: %v", err)
	}
	fmt.Printf("Listening on queue: %s\n", queue)

	// Handle shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Shutting down …")
		receiver.Terminate(5)
		service.Disconnect()
		os.Exit(0)
	}()

	for {
		msg, err := receiver.ReceiveMessage(10 * time.Second)
		if err != nil {
			fmt.Printf("Error receiving message: %v\n", err)
			continue
		}
		if msg != nil {
			payload, _ := msg.GetPayloadAsBytes()
			fmt.Printf("Received message: %s\n", string(payload))
			if err := receiver.Ack(msg); err != nil {
				fmt.Printf("Failed to acknowledge message: %v\n", err)
			}
		} else {
			fmt.Println("Waiting for messages …")
		}
	}
}
