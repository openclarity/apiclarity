#!/bin/bash
TykProxyContainerName="${TYK_PROXY_CONTAINER_NAME:-tyk-gtw}"
TykGatewayDeploymentName="${TYK_GATEWAY_DEPLOYMENT_NAME:-tyk-gtw}"
TykGatewayDeploymentNamespace="${TYK_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
UpstreamTelemetryHostName="${UPSTREAM_TELEMETRY_HOST_NAME:-apiclarity-apiclarity.apiclarity}"
UpstreamTelemetryHTTPPort="${UPSTREAM_TELEMETRY_HTTP_PORT:-9000}"
UpstreamTelemetryTLSPort="${UPSTREAM_TELEMETRY_TLS_PORT:-10443}"
TraceSamplingHostName="${TRACE_SAMPLING_HOST_NAME:-apiclarity-apiclarity.apiclarity:9990}"
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
  kubectl get configmap -n $TykGatewayDeploymentNamespace api-trace-root-ca > /dev/null 2>&1
  if [ $? -eq 0 ]
  then
    kubectl delete configmap -n $TykGatewayDeploymentNamespace api-trace-root-ca
  fi
  kubectl create configmap -n $TykGatewayDeploymentNamespace api-trace-root-ca --from-literal=root-ca.crt="$CERT"

  UpstreamTelemetryHostNameWithPort=$UpstreamTelemetryHostName:$UpstreamTelemetryTLSPort
  deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{ENABLE_TLS}}/$EnableTLS/g" | sed "s/{{TYK_PROXY_CONTAINER_NAME}}/$TykProxyContainerName/g" | sed "s/{{UPSTREAM_TELEMETRY_HOST_NAME}}/$UpstreamTelemetryHostNameWithPort/g" | sed "s/{{TRACE_SOURCE_TOKEN}}/$TraceSourceToken/g" | sed "s/{{TRACE_SAMPLING_ENABLED}}/$TraceSamplingEnabled/g"`
else
  UpstreamTelemetryHostNameWithPort=$UpstreamTelemetryHostName:$UpstreamTelemetryHTTPPort
  # remove certs volume mount from the deployment
  deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{ENABLE_TLS}}/$EnableTLS/g" |  sed '/# {{CERT VOLUME START}}/,/# {{CERT VOLUME END}}/d' | sed '/# {{CERT MOUNT START}}/,/# {{CERT MOUNT END}}/d'  | sed "s/{{TYK_PROXY_CONTAINER_NAME}}/$TykProxyContainerName/g" | sed "s/{{UPSTREAM_TELEMETRY_HOST_NAME}}/$UpstreamTelemetryHostNameWithPort/g" | sed "s/{{TRACE_SOURCE_TOKEN}}/$TraceSourceToken/g" | sed "s/{{TRACE_SAMPLING_ENABLED}}/$TraceSamplingEnabled/g"`
fi

StagedDeployment=$(kubectl get deployments.apps -n ${TykGatewayDeploymentNamespace} ${TykGatewayDeploymentName} -o yaml)
RevisionVersion=$(echo "$StagedDeployment" | grep 'deployment.kubernetes.io/revision: "' | sed -n -e 's/^.*revision: //p' | sed 's/\"//' | sed 's/\"//')
kubectl create configmap -n $TykGatewayDeploymentNamespace tyksnapshot --from-literal data="$RevisionVersion"

kubectl patch deployments.apps -n ${TykGatewayDeploymentNamespace} ${TykGatewayDeploymentName} --patch "$deploymentPatch"
