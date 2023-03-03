#!/bin/bash
KongProxyContainerName="${KONG_PROXY_CONTAINER_NAME:-proxy}"
KongGatewayDeploymentName="${KONG_GATEWAY_DEPLOYMENT_NAME:-kong}"
KongGatewayDeploymentNamespace="${KONG_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
KongGatewayIngressName="${KONG_GATEWAY_INGRESS_NAME:-demo}"
KongGatewayIngressNamespace="${KONG_GATEWAY_INGRESS_NAMESPACE:-default}"

kubectl annotate ingress.networking.k8s.io ${KongGatewayIngressName} -n ${KongGatewayIngressNamespace} konghq.com/plugins-

PreviousRevisionVersion=$(kubectl get configmaps -n $KongGatewayDeploymentNamespace kongsnapshot -o "jsonpath={.data.data}")
kubectl rollout undo deployment/${KongGatewayDeploymentName} -n ${KongGatewayDeploymentNamespace} --to-revision=${PreviousRevisionVersion}

## if configmap exists in namespace, delete it
kubectl get configmap -n $KongGatewayDeploymentNamespace api-trace-root-ca > /dev/null 2>&1
if [ $? -eq 0 ]
then
    kubectl delete configmap -n $KongGatewayDeploymentNamespace api-trace-root-ca
fi

## if configmap exists in namespace, delete it
kubectl get configmap -n $KongGatewayDeploymentNamespace kongsnapshot > /dev/null 2>&1
if [ $? -eq 0 ]
then
    kubectl delete configmap -n $KongGatewayDeploymentNamespace kongsnapshot
fi