# OpenTelemetry Solace Receiver

This project implements an OpenTelemetry receiver for Solace, enabling the reception of telemetry data from Solace Message Brokers and forwarding it to OpenTelemetry-compatible backends.

## Prerequisites

- Go 1.24.2 or higher
- Access to a Solace Message Broker
- OpenTelemetry Collector
- Datadog API Key (for Datadog export)

## Installation

```bash
go get github.com/ThinkportRepo/opentelemetry-solace-otlp
```

## Environment Variables

The project includes a `.env.dist` file as a template for configuration. To set up the environment variables:

1. Copy the `.env.dist` file:
```bash
cp .env.dist .env
```

2. Edit the `.env` file and replace the placeholders with your values:
```bash
# Datadog Configuration
DD_API_KEY=your_datadog_api_key_here
DD_SITE=datadoghq.eu  # For EU region, alternatively datadoghq.com for US region
```

The `.env` file is already listed in `.gitignore` and will not be committed to the repository. The `.env.dist` file serves as a template and documentation for the required environment variables.

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
  - gomod: github.com/ThinkportRepo/opentelemetry-solace-otlp

  # Alternatively, if you want to use a specific version, create a release tag first
  # - gomod: github.com/ThinkportRepo/opentelemetry-solace-otlp v0.0.1

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
  datadog:
    api:
      key: ${DD_API_KEY}  # Read from environment variable
      site: ${DD_SITE}    # Read from environment variable
    host_metadata:
      enabled: true
    metrics:
      endpoint: https://api.${DD_SITE}
    traces:
      endpoint: https://trace.agent.${DD_SITE}
    logs:
      endpoint: https://http-intake.logs.${DD_SITE}

service:
  pipelines:
    traces:
      receivers: [solace]
      exporters: [datadog]
```

## Usage

1. Create the `.env` file with your Datadog API Key and desired Datadog Site
2. Configure the receiver in your OpenTelemetry Collector configuration
3. Start the OpenTelemetry Collector
4. The receiver will now receive telemetry data from your Solace Message Broker

## Starting the Collector

To start the OpenTelemetry Collector, run the following command:

```bash
# Ensure environment variables are loaded
source .env
./otelcol-dev/otelcol-dev --config collector-config.yaml
```

This command starts the collector with the specified configuration file `collector-config.yaml`.

## Debugging the Collector

To debug the collector, you can enable debug output by modifying the `collector-config.yaml` file. Ensure that the debug exporter is enabled by checking the following line in the configuration file:

```yaml
exporters:
  debug:
    verbosity: detailed
```

If you need additional debugging options, you can also adjust the logging configuration to get more detailed information.

### Example Logging Configuration

Add the following line to `collector-config.yaml` to increase the logging level:

```yaml
service:
  telemetry:
    logs:
      level: debug
```

After adjusting the configuration, restart the collector to see the debug output.

## Make Tasks

The project includes a Makefile with the following tasks:

### Build
```bash
make build
```
Builds the OpenTelemetry Collector with the specified configuration in `builder-config.yaml`.

### Start
```bash
make start
```
Starts the OpenTelemetry Collector with the configuration from `collector-config.yaml`.

### Debug
```bash
make debug
```
Starts the OpenTelemetry Collector in debug mode with increased logging level.

### Test Spans
```bash
make test-spans
```
Sends test spans to the OpenTelemetry Collector using otel-cli. This is useful for testing the collector's trace reception capabilities.

#### otel-cli Configuration
The test spans are sent using the following configuration:
- Protocol: gRPC
- Endpoint: 0.0.0.0:4317
- Service Name: test-service
- Span Name: test-span
- Span Kind: client
- Attributes: test.attribute=value

To install otel-cli:
```bash
# macOS
brew install otel-cli

# Linux
curl -L https://github.com/equinix-labs/otel-cli/releases/latest/download/otel-cli-linux-amd64.tar.gz | tar xz
sudo mv otel-cli /usr/local/bin/
```

To send custom spans manually:
```bash
otel-cli span \
  --service "your-service" \
  --name "your-span" \
  --endpoint "0.0.0.0:4317" \
  --protocol grpc \
  --insecure \
  --kind client \
  --attrs "key=value"
```

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please create a Pull Request or open an Issue for suggestions.

## Support

For questions or issues, please create an Issue in this repository. 