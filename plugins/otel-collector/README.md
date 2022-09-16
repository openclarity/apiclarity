# APIClarity HTTP Exporter

| Status                   |                       |
| ------------------------ | --------------------- |
| Stability                | traces [stable]       |
| Supported pipeline types | traces |
| Distributions            | [contrib]     |

Exports traces and/or metrics via HTTP to an [APIClarity](
https://github.com/openclarity/apiclarity/blob/master/plugins/api/swagger.yaml)
endpoint for analysis.

The following settings are required:

- `endpoint` (no default): The target base URL to send data to (e.g.: https://example.com:4318).
  The trace signal will be added to this base URL, i.e. "/api/telemetry" will be appended. 

The following settings can be optionally configured:

- `tls`: see [TLS Configuration Settings](../../config/configtls/README.md) for the full set of available options.
- `timeout` (default = 30s): HTTP request time limit. For details see https://golang.org/pkg/net/http/#Client
- `read_buffer_size` (default = 0): ReadBufferSize for HTTP client.
- `write_buffer_size` (default = 512 * 1024): WriteBufferSize for HTTP client.

Example:

```yaml
exporters:
  apiclarity:
    endpoint: https://example.com:4318/api/telemetry
```

The full list of settings exposed for this exporter are documented [here](./config.go)
with detailed sample configurations [here](./testdata/config.yaml).

[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib

## Using the plugin

In order to use the APIClarity Exporter, you will need to build your own OpenTelemetry Collector with the Exporter included. The instructions to [build a custom collector are here.](https://opentelemetry.io/docs/collector/custom-collector/)

An example builder-config.yaml file including the exporter could be:

```yaml
dist:
  name: otelcol-api
  description: "OTel Collector distribution with APIClarity support"
  output_path: ./otelcol-api

exporters:
  - gomod: "github.com/openclarity/apiclarity/plugins/otel-collector-exporter/exporter"
  - gomod:
      "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/jaegerexporter
      v0.53.0"
  - import: go.opentelemetry.io/collector/exporter/loggingexporter
    gomod: go.opentelemetry.io/collector v0.53.0

receivers:
  - import: go.opentelemetry.io/collector/receiver/otlpreceiver
    gomod: go.opentelemetry.io/collector v0.53.0

processors:
  - import: go.opentelemetry.io/collector/processor/batchprocessor
    gomod: go.opentelemetry.io/collector v0.53.0
```