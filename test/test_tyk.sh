#!/bin/bash

#change values.yaml in order to test your images and change configuration

Registry=="${REGISTRY:-ghcr.io/apiclarity}"
BackendImageTag="${BACKEND_IMAGE_TAG:-v0.8.0}"

## test APIClarity ##

set -euxo pipefail

helm repo update apiclarity

# Deploy Tyk
helm install --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity

cd tyk-oss-k8s-deployment
./launch-tyk.sh

sleep 100

# Port-forward to Tyk & APIClarity
kubectl -n tyk port-forward svc/tyk-svc 8080:8080&
pid1=$!
kubectl port-forward -n apiclarity svc/apiclarity-apiclarity 9999:8080&
pid2=$!

sleep 5

# Run the test suite to generate API calls against Tyk
./generate-apicalls.sh

sleep 10

RESPONSE=$(curl http://localhost:9999/api/apiInventory\?type\=INTERNAL\&page\=1\&pageSize\=50\&sortKey\=name\&sortDir\=DESC)

echo "$RESPONSE" | jq '.items[] | .name'
if echo "$RESPONSE" | grep -q 'jsonplaceholder'; then
  echo "matched"
fi

kill -9 $pid1 $pid2