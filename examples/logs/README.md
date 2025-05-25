# Example: Sending Logs to Solace

This example shows how to send logs to a Solace broker using Go, so they can be received by the opentelemetry-receiver-solace.

## Prerequisites

- Go installed
- Access to a Solace broker
- A configured opentelemetry-receiver-solace

## Usage

1. Set the connection details via environment variables:
   - `SOLACE_HOST`, `SOLACE_VPN`, `SOLACE_USERNAME`, `SOLACE_PASSWORD`, `SOLACE_LOG_TOPIC`
2. Run the example:

```bash
go run main.go
```

The code is based on the pattern from `test/integration/emitter`.
