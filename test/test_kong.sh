#!/bin/bash

#change values.yaml in order to test your images and change configuration

Registry=="${REGISTRY:-ghcr.io/apiclarity}"
BackendImageTag="${BACKEND_IMAGE_TAG:-v0.8.0}"

## test APIClarity ##

set -euxo pipefail

helm repo update apiclarity

#### test kong ##
kubectl create namespace sock-shop

kubectl apply -f https://raw.githubusercontent.com/microservices-demo/microservices-demo/master/deploy/kubernetes/complete-demo.yaml

helm install --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity

# Deploy Kong
kubectl create namespace kong
kubectl apply -f https://bit.ly/kong-ingress-dbless

# Wait for pod to be ready
sleep 15

# Configure Ingress
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: catalogue
  namespace: sock-shop
spec:
  ingressClassName: kong
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: catalogue
            port:
              number: 80
EOF

# Patch kong
cd ../plugins/gateway/kong
KONG_GATEWAY_DEPLOYMENT_NAME=ingress-kong \
KONG_GATEWAY_DEPLOYMENT_NAMESPACE=kong \
KONG_GATEWAY_INGRESS_NAME=catalogue \
KONG_GATEWAY_INGRESS_NAMESPACE=sock-shop \
deploy/deploy.sh

sleep 150

# Get LoadBalacner IP
export PROXY_IP=$(kubectl get -o jsonpath="{.status.loadBalancer.ingress[0].ip}" service -n kong kong-proxy)

# Run Traffic
curl -H 'content-type: application/json' -H 'accept: application/json;charset=UTF-8' $PROXY_IP/catalogue

sleep 2

kubectl port-forward -n apiclarity svc/apiclarity-apiclarity 9999:8080&
pid1=$!
sleep 2

# Check that catalogue service was added to apiclarity
RESPONSE=$(curl http://localhost:9999/api/apiInventory\?type\=INTERNAL\&page\=1\&pageSize\=50\&sortKey\=name\&sortDir\=DESC)

echo "$RESPONSE" | jq '.items[] | .name'
if echo "$RESPONSE" | grep -q 'catalogue'; then
  echo "matched"
fi

kill -9 $pid1
