# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make git curl

# Install OCB
RUN curl -L https://github.com/open-telemetry/opentelemetry-collector-releases/releases/latest/download/ocb_linux_amd64 -o /usr/local/bin/ocb && \
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
COPY --from=builder /app/dist/otelcol-solace /otelcol-solace

# Add labels
LABEL org.opencontainers.image.source="https://github.com/${GITHUB_REPOSITORY}"
LABEL org.opencontainers.image.description="OpenTelemetry Collector with Solace Receiver"
LABEL org.opencontainers.image.licenses="GPL-3.0"

# Use our custom binary as entrypoint
ENTRYPOINT ["/otelcol-solace"] 