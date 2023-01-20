#!/bin/bash

set -eo pipefail

SERVICE_NAME="${SERVICE_NAME:-apiclarity-apiclarity}"
SERVICE_NAMESPACE="${SERVICE_NAMESPACE:-apiclarity}"
TLS_SECRET_NAME="${TLS_SECRET_NAME:-apiclarity-tls}"
TLS_CERT_FILE_NAME="${TLS_CERT_FILE_NAME:-server.crt}"
TLS_KEY_FILE_NAME="${TLS_KEY_FILE_NAME:-server.key}"
ROOT_CERT_CONFIGMAP_NAME="${ROOT_CERT_CONFIGMAP_NAME:-apiclarity-root-ca.crt}"
ROOT_CERT_FILE_NAME="${ROOT_CERT_FILE_NAME:-ca.crt}"

generate_self_signed_certs(){
  openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
    -keyout /tmp/${TLS_KEY_FILE_NAME} -out /tmp/${TLS_CERT_FILE_NAME} -subj "/CN=${SERVICE_NAME}" \
    -extensions san \
    -config <(echo "[req]"; echo "distinguished_name=req";
              echo "[san]"; echo "subjectAltName=DNS:${SERVICE_NAME}.${SERVICE_NAMESPACE},DNS:${SERVICE_NAME}.${SERVICE_NAMESPACE}.svc,DNS:${SERVICE_NAME}.${SERVICE_NAMESPACE}.svc.cluster.local,DNS:${SERVICE_NAME}-external.${SERVICE_NAMESPACE}")

  # Create service TLS secret
  kubectl create secret generic ${TLS_SECRET_NAME} -n ${SERVICE_NAMESPACE} \
  --from-file=/tmp/${TLS_CERT_FILE_NAME} \
  --from-file=/tmp/${TLS_KEY_FILE_NAME}

  # Create root-ca configmap
  cp /tmp/${TLS_CERT_FILE_NAME} /tmp/${ROOT_CERT_FILE_NAME}
  kubectl create configmap ${ROOT_CERT_CONFIGMAP_NAME} -n ${SERVICE_NAMESPACE} --from-file=/tmp/${ROOT_CERT_FILE_NAME}
  echo "self signed certificates for ${SERVICE_NAME}.${SERVICE_NAMESPACE} successfully generated"
}

print_usage() {
  printf "Usage:
  --delete - will delete all certs related k8s objects"
}

while [[ "$1" != "" ]]; do
    case $1 in
        --delete )        DELETE="TRUE"
                          ;;
        -h | --help )          print_usage
                               exit
                               ;;
    *) print_usage
                                  exit 1
  esac
    shift
done

if [ "$DELETE" == "TRUE" ] ; then
  kubectl delete secret generic ${TLS_SECRET_NAME} -n ${SERVICE_NAMESPACE} --ignore-not-found
  kubectl delete configmap ${ROOT_CERT_CONFIGMAP_NAME} -n ${SERVICE_NAMESPACE} --ignore-not-found
else
  generate_self_signed_certs
fi