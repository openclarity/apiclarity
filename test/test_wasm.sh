#!/bin/bash

#change values.yaml in order to test your images and change configuration

Registry=="${REGISTRY:-ghcr.io/apiclarity}"
BackendImageTag="${BACKEND_IMAGE_TAG:-v0.8.0}"

## test APIClarity ##

set -euxo pipefail

helm repo update apiclarity

#### test wasm ##
## pre: need to have istio installed
kubectl create namespace sock-shop
kubectl label namespaces sock-shop istio-injection=enabled

kubectl apply -f https://raw.githubusercontent.com/microservices-demo/microservices-demo/master/deploy/kubernetes/complete-demo.yaml

helm install --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity
cd ../wasm-filters && ./deploy.sh sock-shop
sleep 100

# port-forward to front-end service
kubectl port-forward -n sock-shop svc/front-end 8080:80&
pid1=$!
sleep 2
# create telemetry via request to the frontend
curl http://localhost:8080/catalogue\?page\=1\&size\=6\&tags\=

# port-forward to apiclarity backend
kubectl port-forward -n apiclarity svc/apiclarity-apiclarity 9999:8080&
pid2=$!
sleep 2
# check that catalogue service was added to apiclarity
RESPONSE=$(curl http://localhost:9999/api/apiInventory\?type\=INTERNAL\&page\=1\&pageSize\=50\&sortKey\=name\&sortDir\=DESC)

echo "$RESPONSE" | jq '.items[] | .name'
if echo "$RESPONSE" | grep -q 'catalogue'; then
  echo "matched"
fi

# kill background processes
kill -9 $pid1 $pid2
