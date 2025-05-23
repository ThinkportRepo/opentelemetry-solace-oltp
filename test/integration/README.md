# OTLP Test Sender (Go)

This test program sends OTLP-compliant traces and logs to a Solace queue.

## Prerequisites

- Go 1.21 or higher
- Access to a Solace queue
- Configured `.env` file with Solace credentials

## Installation

1. Initialize the Go module and download dependencies:
```bash
go mod tidy
```

## Configuration

Make sure your `.env` file contains the following variables:
```
SOLACE_ENDPOINT=your_solace_endpoint
```

## Usage

Run the test program:
```bash
go run main.go
```

The program will:
1. Establish a connection to the Solace broker
2. Create a root span with a child span
3. Add events to both spans
4. Send the data in OTLP format to the configured queue

## Verification

You can verify the sent data in your OpenTelemetry Collector, which receives the data from the Solace queue and forwards it to Datadog.

## Program Structure

- `main.go`: Main program with the OTLP sender implementation
- `go.mod`: Go module definition and dependencies
- `.env`: Configuration file for environment variables (not in repository) 