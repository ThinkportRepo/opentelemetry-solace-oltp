```mermaid
graph LR
    subgraph Solace
        S[Solace Message Broker]
    end

    subgraph Custom Collector
        R[Solace Receiver]
        P[Pipeline]
        E[Datadog Exporter]
    end

    subgraph Datadog
        D[Datadog Platform]
        M[Metrics]
        T[Traces]
        L[Logs]
    end

    S -->|Telemetry Data| R
    R -->|Process| P
    P -->|Export| E
    E -->|Metrics| M
    E -->|Traces| T
    E -->|Logs| L

    style Solace fill:#f9f,stroke:#333,stroke-width:2px
    style Custom Collector fill:#bbf,stroke:#333,stroke-width:2px
    style Datadog fill:#bfb,stroke:#333,stroke-width:2px
```

## Description of the data flow

1. **Solace Message Broker**
   - Receives telemetry data from various sources
   - Sends data to the Custom Collector

2. **Custom Collector**
   - **Solace Receiver**: Receives data from the Solace Message Broker
   - **Pipeline**: Processes the received data
   - **Datadog Exporter**: Exports the processed data to Datadog

3. **Datadog Platform**
   - Receives and processes three types of data:
     - Metrics (Metrics)
     - Traces (Traces)
     - Logs (Logs)

The configuration is done via:
- `.env` file for Datadog-specific settings
- `collector-config.yaml` for the OpenTelemetry Collector configuration 