# Kubernetes Deployment Guide: OpenTelemetry Collector with Solace Receiver

This guide explains how to deploy the OpenTelemetry Collector with the Solace Receiver in a Kubernetes cluster.

## Prerequisites

- A running Kubernetes cluster
- `kubectl` configured for your cluster
- Access to a Docker registry (for your custom collector image)
- Datadog API key (if you want to export to Datadog)

## Using .env.dist

The `.env.dist` file in the project root contains example values for all required environment variables. You can use it as a template for local development or for setting up your environment:

```bash
cp ../../.env.dist .env
```

Then, adjust the values in your `.env` file to match your environment. The most important variables for the Solace receiver are:
- `SOLACE_HOST`
- `SOLACE_VPN`
- `SOLACE_USERNAME`
- `SOLACE_PASSWORD`
- `SOLACE_TELEMETRY_QUEUE`

## Quick Start

### 1. Build and Push the Collector Image

Build your custom collector image and push it to your registry:
```bash
docker build -t your-registry/otel-collector-solace:latest .
docker push your-registry/otel-collector-solace:latest
```

### 2. Configure Secrets

Edit `collector-secret.yaml` and set your Datadog API key and site:
```yaml
stringData:
  DD_API_KEY: "your_datadog_api_key_here"
  DD_SITE: "datadoghq.eu"
```
Apply the secret:
```bash
kubectl apply -f collector/k8s/collector-secret.yaml
```

### 3. Apply the ConfigMap

The `collector-config.yaml` contains the OpenTelemetry Collector configuration, including receivers, processors, exporters, and pipelines. You can adjust it as needed.
```bash
kubectl apply -f collector/k8s/collector-config.yaml
```

### 4. Deploy the Collector

The deployment manifest will start the collector with the provided configuration:
```bash
kubectl apply -f collector/k8s/collector-deployment.yaml
```

### 5. Verify the Deployment

Check if the pod is running:
```bash
kubectl get pods -l app=otel-collector
```
View logs:
```bash
kubectl logs -l app=otel-collector
```

## Configuration Details

### ConfigMap (`collector-config.yaml`)
- Defines the OpenTelemetry Collector configuration.
- Adjust receivers, processors, exporters, and service pipelines as needed.
- **To enable the Solace receiver:**
  - Add a `solaceotlp` section under `receivers`.
  - Reference it in the desired pipeline (e.g., `traces`, `logs`).

### Secret (`collector-secret.yaml`)
- Stores sensitive information (e.g., Datadog API key, site).

### Deployment (`collector-deployment.yaml`)
- Defines the Kubernetes deployment for the collector.
- Sets resource limits, ports, volume mounts, and environment variables.
- Mounts the config and secrets into the container.

## Example: Enabling the Solace Receiver

To use the Solace receiver, extend your config as follows:

```yaml
receivers:
  otlp:
    protocols:
      grpc:
      http:
  solaceotlp:
    host: <SOLACE_HOST>
    vpn: <SOLACE_VPN>
    username: <SOLACE_USERNAME>
    password: <SOLACE_PASSWORD>
    queue: <SOLACE_TELEMETRY_QUEUE>

service:
  pipelines:
    traces:
      receivers: [otlp, solaceotlp]
      processors: [batch]
      exporters: [debug, datadog]
```

## Troubleshooting

- Check pod status:
  ```bash
  kubectl describe pod -l app=otel-collector
  ```
- Check collector logs:
  ```bash
  kubectl logs -l app=otel-collector
  ```
- Verify ConfigMap:
  ```bash
  kubectl get configmap otel-collector-config -o yaml
  ```
- Check if secrets are properly mounted:
  ```bash
  kubectl exec -it $(kubectl get pod -l app=otel-collector -o jsonpath='{.items[0].metadata.name}') -- env | grep DD_
  ```

## Files

- `collector/k8s/collector-config.yaml`: Collector configuration (ConfigMap)
- `collector/k8s/collector-secret.yaml`: Secrets for Datadog API key and site
- `collector/k8s/collector-deployment.yaml`: Deployment manifest for the collector

---

For more advanced configuration, see the [OpenTelemetry Collector documentation](https://opentelemetry.io/docs/collector/). 