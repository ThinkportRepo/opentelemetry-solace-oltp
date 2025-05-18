# OpenTelemetry Solace Receiver

This project implements an OpenTelemetry receiver for Solace, enabling the reception of telemetry data from Solace Message Brokers and forwarding it to OpenTelemetry-compatible backends.

## Prerequisites

- Go 1.24.2 or higher
- Access to a Solace Message Broker
- OpenTelemetry Collector

## Installation

```bash
go get github.com/ThinkportRepo/opentelemetry-solace-oltp
```

## Building with OCB

To build a custom OpenTelemetry Collector with this receiver using OCB (OpenTelemetry Collector Builder), follow these steps:

1. Create a `builder-config.yaml` file with the following content:

```yaml
dist:
  name: otelcol-solace
  description: "OpenTelemetry Collector with Solace Receiver"
  output_path: ./dist
  otelcol_version: 0.96.0

receivers:
  # Use the latest version from main branch
  - gomod: github.com/ThinkportRepo/opentelemetry-solace-oltp

  # Alternatively, if you want to use a specific version, create a release tag first
  # - gomod: github.com/ThinkportRepo/opentelemetry-solace-oltp v0.0.1

processors:
  - gomod: go.opentelemetry.io/collector/processor/batchprocessor v0.96.0

exporters:
  - gomod: go.opentelemetry.io/collector/exporter/otlpexporter v0.96.0
  - gomod: go.opentelemetry.io/collector/exporter/loggingexporter v0.96.0

extensions:
  - gomod: go.opentelemetry.io/collector/extension/healthcheckextension v0.96.0
```

2. Build the collector using OCB:

```bash
ocb --config builder-config.yaml
```

The resulting binary will be available in the `./dist` directory.

Note: If you want to use a specific version of the receiver, you need to:
1. Create a release tag in the repository (e.g., v0.0.1)
2. Update the version in the builder-config.yaml accordingly

## Configuration

Configuration is done through the OpenTelemetry Collector configuration file. Example:

```yaml
receivers:
  solace:
    endpoint: "tcp://localhost:5672"
    queue: "telemetry-queue"
    username: "default"
    password: "default"

exporters:
  otlp:
    endpoint: "localhost:4317"
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [solace]
      exporters: [otlp]
```

## Usage

1. Configure the receiver in your OpenTelemetry Collector configuration
2. Start the OpenTelemetry Collector
3. The receiver will now receive telemetry data from your Solace Message Broker

## License

[Add license information here]

## Contributing

Contributions are welcome! Please create a Pull Request or open an Issue for suggestions.

## Support

For questions or issues, please create an Issue in this repository.

## Starten des Collectors

Um den OpenTelemetry Collector zu starten, führen Sie den folgenden Befehl aus:

```bash
./otelcol-dev/otelcol-dev --config collector-config.yaml
```

Dieser Befehl startet den Collector mit der angegebenen Konfigurationsdatei `collector-config.yaml`.

## Debuggen des Collectors

Um den Collector zu debuggen, können Sie die Debug-Ausgabe aktivieren, indem Sie die Konfigurationsdatei `collector-config.yaml` anpassen. Stellen Sie sicher, dass der Debug-Exporter aktiviert ist, indem Sie die folgende Zeile in der Konfigurationsdatei überprüfen:

```yaml
exporters:
  debug:
    verbosity: detailed
```

Wenn Sie weitere Debugging-Optionen benötigen, können Sie auch die Logging-Konfiguration anpassen, um detailliertere Informationen zu erhalten. 

### Beispiel für die Logging-Konfiguration

Fügen Sie die folgende Zeile in die `collector-config.yaml` ein, um die Logging-Stufe zu erhöhen:

```yaml
service:
  telemetry:
    logs:
      level: debug
```

Nachdem Sie die Konfiguration angepasst haben, starten Sie den Collector erneut, um die Debug-Ausgabe zu sehen. 