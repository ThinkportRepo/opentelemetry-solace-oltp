# Solace OTLP Receiver

[![Coverage](https://raw.githubusercontent.com/ThinkportRepo/opentelemetry-solace-otlp/main/coverage.json)](https://github.com/ThinkportRepo/opentelemetry-solace-otlp)


This receiver enables receiving OpenTelemetry data via Solace Message Broker.

## Status

This receiver is in **alpha** status. The following features are supported:

### Supported Data Types

| Data Type | Supported |
| --------- | --------- |
| Metrics   | ❌        |
| Logs      | ✅        |
| Traces    | ✅        |

## Configuration

The receiver supports the following configuration options:

```yaml
receivers:
  solaceotlp:
    endpoint: "tcp://localhost:55555" # Solace Message Broker Endpoint
    queue: "otel-traces" # Queue name for traces
    username: "default" # Solace username
    password: "default" # Solace password
    vpn: "default" # Solace VPN name
```

### Configuration Fields

| Field      | Description                                  | Default                 |
| ---------- | -------------------------------------------- | ----------------------- |
| `endpoint` | The endpoint of the Solace Message Broker    | `tcp://localhost:55555` |
| `queue`    | The name of the queue to receive traces from | `otel-traces`           |
| `username` | The username for the Solace connection       | `default`               |
| `password` | The password for the Solace connection       | `default`               |
| `vpn`      | The VPN name for the Solace connection       | `default`               |

## Features

- Receiving OpenTelemetry traces via Solace Message Broker
- Support for various Solace queue types
- Automatic message acknowledgment
- Configurable connection parameters

## Example Configuration

Here is an example of a complete configuration:

```yaml
receivers:
  solaceotlp:
    endpoint: "tcp://solace-broker:55555"
    queue: "otel-traces"
    username: "otel-user"
    password: "otel-password"

processors:
  batch:

exporters:
  otlp:
    endpoint: "otel-collector:4317"
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [solaceotlp]
      processors: [batch]
      exporters: [otlp]
```

## Prerequisites

- Solace Message Broker (Version 10.x or higher)
- OpenTelemetry Collector
- Go 1.21 or higher

## Installation

The receiver can be installed via the OpenTelemetry Collector Builder:

```bash
go install go.opentelemetry.io/collector/cmd/builder@latest
```

Then add the receiver to your builder configuration:

```yaml
receivers:
  - gomod: github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver v0.0.1
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

## License

Apache License 2.0
