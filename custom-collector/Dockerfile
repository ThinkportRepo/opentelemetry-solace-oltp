# Build stage
FROM golang:1.24.2-alpine AS builder

ARG OCB_VERSION=v0.126.0

# Install build dependencies
RUN apk add --no-cache make git curl

# Download OCB binary for Linux
RUN curl --proto '=https' --tlsv1.2 -fL -o /usr/local/bin/ocb \
    https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/cmd%2Fbuilder%2Fv0.126.0/ocb_0.126.0_linux_amd64 && \
    chmod +x /usr/local/bin/ocb

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the collector
RUN make build

# Final stage
FROM otel/opentelemetry-collector-contrib:0.126.0

# Copy the custom binary
COPY --from=builder /app/otelcol-dev/otelcol-dev /otelcol-solace

# Add labels
LABEL org.opencontainers.image.source="https://github.com/ThinkportRepo/opentelemetry-receiver-solace"
LABEL org.opencontainers.image.description="OpenTelemetry Collector with Solace Receiver"
LABEL org.opencontainers.image.licenses="GPL-3.0"

# Use our custom binary as entrypoint
ENTRYPOINT ["/otelcol-solace"]
