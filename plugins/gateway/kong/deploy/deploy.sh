#!/bin/bash
KongProxyContainerName="${KONG_PROXY_CONTAINER_NAME:-proxy}"
KongGatewayDeploymentName="${KONG_GATEWAY_DEPLOYMENT_NAME:-kong}"
KongGatewayIngressName="${KONG_GATEWAY_INGRESS_NAME:-demo}"

kubectl apply -f kongPlugin.yaml

deploymentPatch=`cat "patch-deployment.yaml" | sed "s/{{KONG_PROXY_CONTAINER_NAME}}/$KongProxyContainerName/g"`

kubectl patch deployments.apps ${KongGatewayDeploymentName} --patch "$deploymentPatch"
kubectl patch ingresses.networking.k8s.io ${KongGatewayIngressName} --patch "$(cat patch-ingress.yaml)"
