#!/bin/bash
KongProxyContainerName="${KONG_PROXY_CONTAINER_NAME:-proxy}"
KongGatewayDeploymentName="${KONG_GATEWAY_DEPLOYMENT_NAME:-kong}"
KongGatewayDeploymentNamespace="${KONG_GATEWAY_DEPLOYMENT_NAMESPACE:-default}"
KongGatewayIngressName="${KONG_GATEWAY_INGRESS_NAME:-demo}"
KongGatewayIngressNamespace="${KONG_GATEWAY_INGRESS_NAMESPACE:-default}"
UpstreamTelemetryHostName="${UPSTREAM_TELEMETRY_HOST_NAME:-nats-proxy.portshift.svc.cluster.local:1323}"
EnableTLS="${ENABLE_TLS:-false}"

CERT_DIR="/tmp"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cat "${DIR}/kongPlugin.yaml" | sed "s/{{UPSTREAM_TELEMETRY_HOST_NAME}}/$UpstreamTelemetryHostName/g" | kubectl -n ${KongGatewayIngressNamespace} apply -f -

deploymentPatch=`cat "${DIR}/patch-deployment.yaml" | sed "s/{{KONG_PROXY_CONTAINER_NAME}}/$KongProxyContainerName/g"`

kubectl patch deployments.apps -n ${KongGatewayDeploymentNamespace} ${KongGatewayDeploymentName} --patch "$deploymentPatch"
kubectl patch ingresses.networking.k8s.io -n ${KongGatewayIngressNamespace} ${KongGatewayIngressName} --patch "$(cat ${DIR}/patch-ingress.yaml)"

generate_self_signed_certs

"${CERT_DIR}"/root-cert.pem
"${CERT_DIR}"/ca-cert.pem
"${CERT_DIR}"/ca-key.pem
"${CERT_DIR}"/cert-chain.pem

kubectl create secret tls client-secret \
  --cert=path/to/cert/file \
  --key=path/to/key/file

kubectl create secret tls server-secret \
  --cert=path/to/cert/file \
  --key=path/to/key/file

generate_self_signed_certs(){
  ERR=$("$DIR"/certs_gen.sh -c $CERT_DIR)
  if [[ $? -ne 0 ]]
  then
    echo "failed to generate certificates"
    echo ${ERR}
    exit 1
  fi
  echo "self signed certificates successfully generated"
}
