#!/bin/bash
kubectl delete ingresses.networking.k8s.io -n sock-shop catalogue
kubectl delete kongplugins.configuration.konghq.com -n sock-shop api-traces-plugin
helm uninstall -n apiclarity apiclarity 2>&1 || true
kubectl delete ns sock-shop 2>&1 || true

kubectl delete ns kong 2>&1 || true
kubectl delete ingresses.networking.k8s.io -n sock-shop catalogue 2>&1 || true

kubectl delete ns tyk 2>&1 || true
