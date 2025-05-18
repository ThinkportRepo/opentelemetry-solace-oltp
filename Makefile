.PHONY: build start debug

# Include environment variables from .env file
include .env
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