#!/bin/bash
TykProxyContainerName="${TYK_PROXY_CONTAINER_NAME:-tyk-gtw}"
TykGatewayDeploymentName="${TYK_GATEWAY_DEPLOYMENT_NAME:-tyk-gtw}"
TykGatewayDeploymentNamespace="${TYK_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
UpstreamTelemetryAddress="${UPSTREAM_TELEMETRY_ADDRESS:-apiclarity-apiclarity.apiclarity:9000}"
TraceSamplingAddress="${TRACE_SAMPLING_ADDRESS:-apiclarity-apiclarity.apiclarity:9990}"
TraceSamplingEnabled="${TRACE_SAMPLING_ENABLED:-false}"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{TYK_PROXY_CONTAINER_NAME}}/$TykProxyContainerName/g" | sed "s/{{UPSTREAM_TELEMETRY_ADDRESS}}/$UpstreamTelemetryAddress/g" | sed "s/{{TRACE_SAMPLING_ADDRESS}}/$TraceSamplingAddress/g" | sed "s/{{TRACE_SAMPLING_ENABLED}}/$TraceSamplingEnabled/g"`

kubectl patch deployments.apps -n ${TykGatewayDeploymentNamespace} ${TykGatewayDeploymentName} --patch "$deploymentPatch"
