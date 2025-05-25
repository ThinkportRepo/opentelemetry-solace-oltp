# OpenTelemetry Solace OTLP

This project is a custom OpenTelemetry Collector with a receiver for Solace with OTLP data. 
It supports logs and traces.

## Content

1. [Opentelemetry Solace Receiver](./receiver/solaceotlpreceiver/README.md)
2. [Build your custom OpenTelemetry Collector with the solace receiver](./custom-collector/README.md)

## Supported Message Formats

The receiver supports the following message formats for both traces and logs:

### Traces
1. Direct Protobuf OTLP Traces
   - Raw protobuf-encoded OTLP trace data
   - Example: `\u001f\u0000\u0000\u0000...`

2. Base64-encoded Protobuf OTLP Traces
   - Base64-encoded protobuf OTLP trace data
   - Example: `eyJ0cmFjZV9pZCI6IjEyMzQ1Njc4OTBhYmNkZWYiLCJzcGFuX2lkIjoiMTIzNDU2Nzg5MGFiY2RlZiJ9`

3. Base64-encoded JSON Traces
   - Base64-encoded JSON trace data
   - Example: `eyJ0cmFjZV9pZCI6IjEyMzQ1Njc4OTBhYmNkZWYiLCJzcGFuX2lkIjoiMTIzNDU2Nzg5MGFiY2RlZiJ9`

4. Direct JSON Traces
   - Raw JSON trace data
   - Example: `{"trace_id":"1234567890abcdef","span_id":"1234567890abcdef"}`

### Logs
1. Direct Protobuf OTLP Logs
   - Raw protobuf-encoded OTLP log data
   - Example: `\u001f\u0000\u0000\u0000...`

2. Base64-encoded Protobuf OTLP Logs
   - Base64-encoded protobuf OTLP log data
   - Example: `eyJzZXZlcml0eV9udW1iZXIiOjksInNldmVyaXR5X3RleHQiOiJJTkZPIiwibWVzc2FnZSI6IlRlc3QgbG9nIn0=`

3. Base64-encoded JSON Logs
   - Base64-encoded JSON log data
   - Example: `eyJzZXZlcml0eV9udW1iZXIiOjksInNldmVyaXR5X3RleHQiOiJJTkZPIiwibWVzc2FnZSI6IlRlc3QgbG9nIn0=`

4. Direct JSON Logs
   - Raw JSON log data
   - Example: `{"severity_number":9,"severity_text":"INFO","message":"Test log"}`

The receiver will automatically detect and parse the appropriate format based on the message content.

## Contributing

Please read [CONTRIBUTE.md](CONTRIBUTE.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the GNU GENERAL PUBLIC LICENSE License - see the [LICENSE](LICENSE) file for details. 
