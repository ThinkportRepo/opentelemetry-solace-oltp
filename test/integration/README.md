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
SOLACE_HOST=your_solace_host
```

## Solace Configuration

Before running the test program, ensure the following settings are configured in your Solace broker:

1. **Queue Creation**:
   - Create a queue named `otlp-traces` (or your preferred name)
   - Set the queue type to "Standard"
   - Enable message delivery guarantees (at least once)

2. **Queue Permissions**:
   - Grant the following permissions to your client username:
     - `consume`
     - `send`
     - `read`
     - `write`

3. **Message VPN Settings**:
   - Ensure the Message VPN is configured to allow client connections
   - Configure appropriate authentication settings (username/password or client certificate)
   - Set appropriate message size limits (recommended: at least 1MB)

4. **Client Profile**:
   - Create or modify a client profile with:
     - Enabled client connections
     - Appropriate message size limits
     - Required authentication methods

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