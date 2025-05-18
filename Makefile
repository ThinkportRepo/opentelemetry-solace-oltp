.PHONY: build start debug docker-build docker-build-local docker-push

# Include environment variables from .env file if it exists
-include .env
export

# Build the OpenTelemetry Collector
build:
	ocb --config builder-config.yaml

# Start the OpenTelemetry Collector
start:
	./otelcol-dev/otelcol-dev --config collector-config.yaml

# Build and start the OpenTelemetry Collector
rebuild:
	make build
	make start

# Debug the OpenTelemetry Collector
debug:
	./otelcol-dev/otelcol-dev --config collector-config.yaml --log-level=debug

# Build Docker image for Linux AMD64 (default)
docker-build:
	docker build -t ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest .

# Build Docker image for Mac ARM64 (local development)
docker-build-local:
	docker build -t ghcr.io/thinkportrepo/opentelemetry-receiver-solace:local .

docker-push:
	docker push ghcr.io/thinkportrepo/opentelemetry-receiver-solace:latest 