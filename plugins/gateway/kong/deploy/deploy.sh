#!/bin/bash
KongProxyContainerName="${KONG_PROXY_CONTAINER_NAME:-proxy}"
KongGatewayDeploymentName="${KONG_GATEWAY_DEPLOYMENT_NAME:-kong}"
KongGatewayDeploymentNamespace="${KONG_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
KongGatewayIngressName="${KONG_GATEWAY_INGRESS_NAME:-demo}"
KongGatewayIngressNamespace="${KONG_GATEWAY_INGRESS_NAMESPACE:-default}"
UpstreamTelemetryHostName="${UPSTREAM_TELEMETRY_HOST_NAME:-apiclarity.apiclarity:9000}"

cat "kongPlugin.yaml" | sed "s/{{UPSTREAM_TELEMETRY_HOST_NAME}}/$UpstreamTelemetryHostName/g" | kubectl -n ${KongGatewayIngressNamespace} apply -f -

deploymentPatch=`cat "patch-deployment.yaml" | sed "s/{{KONG_PROXY_CONTAINER_NAME}}/$KongProxyContainerName/g"`

kubectl patch deployments.apps -n ${KongGatewayDeploymentNamespace} ${KongGatewayDeploymentName} --patch "$deploymentPatch"
kubectl patch ingresses.networking.k8s.io -n ${KongGatewayIngressNamespace} ${KongGatewayIngressName} --patch "$(cat patch-ingress.yaml)"
