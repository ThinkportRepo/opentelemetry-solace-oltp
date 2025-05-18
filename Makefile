.PHONY: build start debug

# Build the OpenTelemetry Collector
build:
	ocb --config builder-config.yaml

# Start the OpenTelemetry Collector
start:
	./otelcol-dev/otelcol-dev --config collector-config.yaml

# Debug the OpenTelemetry Collector
debug:
	./otelcol-dev/otelcol-dev --config collector-config.yaml --log-level=debug 