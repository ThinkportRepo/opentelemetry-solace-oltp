# Contributing

## How to contribute

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Make your changes.
4. Submit a pull request.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

## Go-related Information

### Prerequisites

- Go 1.24.2 or higher
- Access to a Solace Message Broker
- OpenTelemetry Collector
- Datadog API Key (for Datadog export)

### Installation

```sh
go get github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver
```

### Usage

```go
import "github.com/ThinkportRepo/opentelemetry-solace-otlp/receiver/solaceotlpreceiver"
```

## Build Process

### Building with OCB

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

## Automated Releases with release-please

This project uses [release-please](https://github.com/googleapis/release-please) to automate the release process. release-please analyzes commit messages (following the Conventional Commits specification) to determine the type of changes (features, fixes, breaking changes) and automatically manages versioning and changelog updates.

How it works:

- When changes are merged into the main branch, release-please creates a pull request proposing a new release. This PR includes:
  - An updated version in relevant files (according to semantic versioning)
  - An updated `CHANGELOG.md` summarizing the changes since the last release
- Once the release PR is merged, release-please creates a new GitHub release, tags the commit, and publishes the release notes.
- The configuration for release-please can be found in `release-please-config.json`.

This automation helps to standardize and simplify the release workflow for all contributors.

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE License - see the [LICENSE](LICENSE) file for details.
