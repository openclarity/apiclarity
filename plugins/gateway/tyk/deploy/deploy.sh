#!/bin/bash
TykProxyContainerName="${TYK_PROXY_CONTAINER_NAME:-tyk-gtw}"
TykGatewayDeploymentName="${TYK_GATEWAY_DEPLOYMENT_NAME:-tyk-gtw}"
TykGatewayDeploymentNamespace="${TYK_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"


deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{TYK_PROXY_CONTAINER_NAME}}/$TykProxyContainerName/g"`

kubectl patch deployments.apps -n ${TykGatewayDeploymentNamespace} ${TykGatewayDeploymentName} --patch "$deploymentPatch"
