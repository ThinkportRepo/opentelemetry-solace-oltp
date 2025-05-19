# OpenTelemetry Solace OTLP

[![Coverage](https://raw.githubusercontent.com/edeka-digital/opentelemetry-receiver-solace/main/coverage.json)](https://github.com/edeka-digital/opentelemetry-receiver-solace)

This project is a custom OpenTelemetry Collector with a receiver for Solace OTLP.

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

## Contributing

Please read [CONTRIBUTE.md](CONTRIBUTE.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 