# Kubernetes Deployment Guide

This guide explains how to deploy the OpenTelemetry Collector with Solace Receiver to a Kubernetes cluster.

## Prerequisites

- Kubernetes cluster
- kubectl configured
- Docker registry access
- Datadog API key

## Steps

1. Build and push the Docker image:
```bash
# Build the image
docker build -t your-registry/otel-collector-solace:latest .

# Push to your registry
docker push your-registry/otel-collector-solace:latest
```

2. Update the Datadog API key in the secret:
```bash
# Edit the secret file
kubectl edit secret otel-collector-secrets
```

3. Apply the Kubernetes manifests:
```bash
# Create the ConfigMap
kubectl apply -f collector-config.yaml

# Create the Secret
kubectl apply -f collector-secret.yaml

# Deploy the Collector
kubectl apply -f collector-deployment.yaml
```

4. Verify the deployment:
```bash
# Check if the pod is running
kubectl get pods -l app=otel-collector

# Check the logs
kubectl logs -l app=otel-collector
```

## Configuration

### ConfigMap
The `collector-config.yaml` ConfigMap contains the OpenTelemetry Collector configuration. You can modify it to adjust:
- Receivers
- Processors
- Exporters
- Service pipelines

### Secret
The `collector-secret.yaml` Secret contains sensitive information:
- Datadog API key
- Datadog site configuration

### Deployment
The `collector-deployment.yaml` defines the Kubernetes deployment with:
- Resource limits and requests
- Port configurations
- Volume mounts
- Environment variables

## Troubleshooting

1. Check pod status:
```bash
kubectl describe pod -l app=otel-collector
```

2. Check collector logs:
```bash
kubectl logs -l app=otel-collector
```

3. Verify ConfigMap:
```bash
kubectl get configmap otel-collector-config -o yaml
```

4. Check if secrets are properly mounted:
```bash
kubectl exec -it $(kubectl get pod -l app=otel-collector -o jsonpath='{.items[0].metadata.name}') -- env | grep DD_
``` 