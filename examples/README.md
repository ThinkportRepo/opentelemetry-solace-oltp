# Examples: Sending Logs and Traces to Solace

This directory contains Go examples for sending logs and traces to Solace, so they can be received by the opentelemetry-receiver-solace.

## Prerequisites

- Go installed
- Access to a Solace broker
- The opentelemetry-receiver-solace is running and properly configured

## Examples

- [logs/](logs/) – Beispiel zum Senden von Logs
- [traces/](traces/) – Beispiel zum Senden von Traces
- [k8s/](k8s/) – Beispiel für die Bereitstellung des OpenTelemetry Collector mit Solace Receiver in Kubernetes

Each subdirectory contains its own README and a runnable Go example based on the integration/emitter pattern.
