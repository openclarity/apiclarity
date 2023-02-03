#!/bin/bash
KongProxyContainerName="${KONG_PROXY_CONTAINER_NAME:-proxy}"
KongGatewayDeploymentName="${KONG_GATEWAY_DEPLOYMENT_NAME:-kong}"
KongGatewayDeploymentNamespace="${KONG_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
KongGatewayIngressName="${KONG_GATEWAY_INGRESS_NAME:-demo}"
KongGatewayIngressNamespace="${KONG_GATEWAY_INGRESS_NAMESPACE:-default}"
UpstreamTelemetryHostName="${UPSTREAM_TELEMETRY_HOST_NAME:-apiclarity-apiclarity.apiclarity}"
UpstreamTelemetryHTTPPort="${UPSTREAM_TELEMETRY_HTTP_PORT:-9000}"
UpstreamTelemetryTLSPort="${UPSTREAM_TELEMETRY_TLS_PORT:-9443}"
TraceSamplingEnabled="${TRACE_SAMPLING_ENABLED:-false}"
EnableTLS="${ENABLE_TLS:-false}"
RootCertConfigMapName="${ROOT_CERT_CONFIGMAP_NAME:-apiclarity-root-ca.crt}"
RootCertConfigMapNamespace="${ROOT_CERT_CONFIGMAP_NAMESPACE:-apiclarity}"
RootCertFileName="${ROOT_CERT_FILE_NAME:-ca.crt}"
RootCertFileNameEscaped=$(echo ${RootCertFileName} | sed "s/[.]/\\\&/g")
TraceSourceToken="$TRACE_SOURCE_TOKEN"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$EnableTLS" == "true" ]
then
  kubectl get configmap -n $RootCertConfigMapNamespace $RootCertConfigMapName
  if [[ $? -ne 0 ]]
  then
    echo "Root CA cert config map ($RootCertConfigMapName) is missing from $RootCertConfigMapNamespace namespace, consider setting ENABLE_TLS to false"
    exit 1
  fi
  # copy root ca configmap from provided namespace
  CERT=$(kubectl get configmap -n $RootCertConfigMapNamespace $RootCertConfigMapName -o jsonpath="{.data.${RootCertFileNameEscaped}}")
  ## if configmap already exists in namespace, delete it
  kubectl get configmap -n $KongGatewayDeploymentNamespace api-trace-root-ca > /dev/null 2>&1
  if [ $? -eq 0 ]
  then
    kubectl delete configmap -n $KongGatewayDeploymentNamespace api-trace-root-ca
  fi
  kubectl create configmap -n $KongGatewayDeploymentNamespace api-trace-root-ca --from-literal=root-ca.crt="$CERT"

  UpstreamTelemetryHostNameWithPort=$UpstreamTelemetryHostName:$UpstreamTelemetryTLSPort
  deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{KONG_PROXY_CONTAINER_NAME}}/$KongProxyContainerName/g"`
else
  UpstreamTelemetryHostNameWithPort=$UpstreamTelemetryHostName:$UpstreamTelemetryHTTPPort
  # remove certs volume mount from the deployment
  deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed '/# {{CERT VOLUME START}}/,/# {{CERT VOLUME END}}/d' | sed '/# {{CERT MOUNT START}}/,/# {{CERT MOUNT END}}/d' | sed "s/{{KONG_PROXY_CONTAINER_NAME}}/$KongProxyContainerName/g"`
fi

cat "${DIR}/kongPlugin.yaml" | sed "s/{{TRACE_SOURCE_TOKEN}}/$TraceSourceToken/g" | sed "s/{{UPSTREAM_TELEMETRY_HOST}}/$UpstreamTelemetryHostNameWithPort/g" | sed "s/{{TRACE_SAMPLING_ENABLED}}/$TraceSamplingEnabled/g" | sed "s/{{ENABLE_TLS}}/$EnableTLS/g" | kubectl -n ${KongGatewayIngressNamespace} apply -f -

kubectl patch deployments.apps -n ${KongGatewayDeploymentNamespace} ${KongGatewayDeploymentName} --patch "$deploymentPatch"
kubectl patch ingresses.networking.k8s.io -n ${KongGatewayIngressNamespace} ${KongGatewayIngressName} --patch "$(cat ${DIR}/patch-ingress.yaml)"
