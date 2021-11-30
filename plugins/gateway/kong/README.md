APIClarity Kong gateway plugin

Prerequisuts:
Make sure thst kong gateway is installed in your cluster, and that he is configured with an ingress resource.
For quick installation:
# Deploy Kong
kubectl create namespace kong
kubectl apply -f https://bit.ly/kong-ingress-dbless

# Wait for pod to be ready
watch kubectl get pods -n kong

# Configure Ingress
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: <name>
  namespace: <namespace>
spec:
  ingressClassName: kong
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: <service-name>
            port:
              number: 80
EOF

Refer to the documentation for more details: https://docs.konghq.com/gateway/2.6.x/install-and-run/kubernetes/

Installation using pre-built image

```shell
KONG_PROXY_CONTAINER_NAME=<name> KONG_GATEWAY_DEPLOYMENT_NAME=<name> KONG_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> KONG_GATEWAY_INGRESS_NAME=<name> KONG_GATEWAY_INGRESS_NAMESPACE=<namespace> UPSTREAM_TELEMETRY_HOST_NAME=<telemetry service address> ./deploy/deploy.sh
```

Where:

KONG_PROXY_CONTAINER_NAME - the name of the proxy container in Kong gateway (default to proxy)

KONG_GATEWAY_DEPLOYMENT_NAME - the name of the Kong gateway deployment to be patched

KONG_GATEWAY_DEPLOYMENT_NAMESPACE - the namespace of the Kong gateway deployment to be patched

KONG_GATEWAY_INGRESS_NAME - the name of the ingress resource to be patched

KONG_GATEWAY_INGRESS_NAMESPACE - the namespace of the ingress resource to be patched

UPSTREAM_TELEMETRY_HOST_NAME - The name of the telemetry service (defaults to apiclarity-apiclarity.apiclarity:9000)

Once the plugin is deployed, traces will be sent to APIClarity to start learning specs.
  


