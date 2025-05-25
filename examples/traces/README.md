# Example: Sending Traces to Solace

This example shows how to send traces to a Solace broker using Go, so they can be received by the opentelemetry-receiver-solace.

## Prerequisites

- Go installed
- Access to a Solace broker
- A configured opentelemetry-receiver-solace

## Using .env.dist

The `.env.dist` file contains example values for all required environment variables. You can use it as a template:

```bash
cp ../../.env.dist .env
```

Then, adjust the values in your `.env` file to match your environment. The most important variables are:

- `SOLACE_HOST`
- `SOLACE_VPN`
- `SOLACE_USERNAME`
- `SOLACE_PASSWORD`
- `SOLACE_TRACE_TOPIC`

## Usage

1. Set the connection details via environment variables:
   - `SOLACE_HOST`, `SOLACE_VPN`, `SOLACE_USERNAME`, `SOLACE_PASSWORD`, `SOLACE_TRACE_TOPIC`
2. Run the example:

```bash
go run main.go
```

The code is based on the pattern from `test/integration/emitter`.
