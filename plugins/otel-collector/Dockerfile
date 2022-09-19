FROM alpine:3.14 as certs
RUN apk --update add ca-certificates

FROM golang:1.18-alpine3.14 AS otelcol
ADD https://github.com/open-telemetry/opentelemetry-collector/releases/download/v0.59.0/ocb_0.59.0_linux_amd64 /usr/local/bin/ocb
RUN chmod 755 /usr/local/bin/ocb

COPY otelcol-build /otelcol-build/
COPY apiclarityexporter /apiclarityexporter/
WORKDIR /otelcol-build
RUN /usr/local/bin/ocb --config ./otelcol-builder.yaml
# Note that this shouldn't be necessary, but in some cases the file seems to be
# copied with the execute bit lost (see #1317)
RUN chmod 755 /otelcol-build/otelcol

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=otelcol /otelcol-build/otelcol /
COPY configs/otel-collector-config.yaml /etc/otelcol/config.yaml
ENTRYPOINT ["/otelcol"]
CMD ["--config", "/etc/otelcol/config.yaml"]
EXPOSE 4317 8888 55678 55679