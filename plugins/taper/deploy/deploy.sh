#!/bin/bash

UpstreamTelemetryHostName="${UPSTREAM_TELEMETRY_HOST_NAME:-apiclarity-apiclarity.apiclarity:9000}"
NamespacesToTap="${NAMESPACES_TO_TAP:-default}"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cat "${DIR}/gopacket-ds.yml" | sed "s/{{NAMESPACES_TO_TAP}}/\"$NamespacesToTap\"/g" | sed "s/{{UPSTREAM_TELEMETRY_HOST_NAME}}/$UpstreamTelemetryHostName/g" | kubectl -n apiclarity apply -f -
