#!/bin/bash
TykProxyContainerName="${TYK_PROXY_CONTAINER_NAME:-tyk-gtw}"
TykGatewayDeploymentName="${TYK_GATEWAY_DEPLOYMENT_NAME:-tyk-gtw}"
TykGatewayDeploymentNamespace="${TYK_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"

PreviousRevisionVersion=$(kubectl get configmaps -n $TykGatewayDeploymentNamespace tyksnapshot -o "jsonpath={.data.data}")
kubectl rollout undo deployment/${TykGatewayDeploymentName} -n ${TykGatewayDeploymentNamespace} --to-revision=${PreviousRevisionVersion}

## if configmap already exists in namespace, delete it
kubectl get configmap -n $TykGatewayDeploymentNamespace api-trace-root-ca > /dev/null 2>&1
if [ $? -eq 0 ]
then
    kubectl delete configmap -n $TykGatewayDeploymentNamespace api-trace-root-ca
fi

## if configmap already exists in namespace, delete it
kubectl get configmap -n $TykGatewayDeploymentNamespace tyksnapshot > /dev/null 2>&1
if [ $? -eq 0 ]
then
    kubectl delete configmap -n $TykGatewayDeploymentNamespace tyksnapshot
fi