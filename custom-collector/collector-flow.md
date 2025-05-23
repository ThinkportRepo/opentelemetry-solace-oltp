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

## Beschreibung des Datenflusses

1. **Solace Message Broker**
   - Empfängt Telemetriedaten von verschiedenen Quellen
   - Sendet Daten an den Custom Collector

2. **Custom Collector**
   - **Solace Receiver**: Empfängt Daten vom Solace Message Broker
   - **Pipeline**: Verarbeitet die empfangenen Daten
   - **Datadog Exporter**: Exportiert die verarbeiteten Daten an Datadog

3. **Datadog Platform**
   - Empfängt und verarbeitet drei Arten von Daten:
     - Metrics (Metriken)
     - Traces (Spuren)
     - Logs (Protokolle)

Die Konfiguration erfolgt über:
- `.env` Datei für Datadog-spezifische Einstellungen
- `collector-config.yaml` für die OpenTelemetry Collector Konfiguration 