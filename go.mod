module github.com/ThinkportRepo/opentelemetry-solace-otlp

go 1.23.0

toolchain go1.24.2

require (
	github.com/joho/godotenv v1.5.1
	go.opentelemetry.io/proto/otlp v1.6.0
	google.golang.org/protobuf v1.36.6
	solace.dev/go/messaging v1.3.0
)

require github.com/google/go-cmp v0.7.0 // indirect
