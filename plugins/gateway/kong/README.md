# APIClarity Kong gateway plugin

## _Prerequisuts:_ 

Make sure thst kong gateway is installed in your cluster, and that he is configured with an ingress resource.

### For quick installation:

### Deploy Kong
```sh
kubectl create namespace kong
kubectl apply -f https://bit.ly/kong-ingress-dbless
```

### Wait for pod to be ready
```sh
watch kubectl get pods -n kong
```

### Configure Ingress
```sh
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
```

Refer to the documentation for more details: https://docs.konghq.com/gateway/2.6.x/install-and-run/kubernetes/

  
## _Deploy The Plugin:_
  
### Run:
  
```shell
KONG_PROXY_CONTAINER_NAME=<name> KONG_GATEWAY_DEPLOYMENT_NAME=<name> KONG_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> KONG_GATEWAY_INGRESS_NAME=<name> KONG_GATEWAY_INGRESS_NAMESPACE=<namespace> UPSTREAM_TELEMETRY_HOST_NAME=<telemetry service address> ./deploy/deploy.sh
```

Where:

## _KONG_PROXY_CONTAINER_NAME_ - the name of the proxy container in Kong gateway deployment (default to proxy)

## _KONG_GATEWAY_DEPLOYMENT_NAME_ - the name of the Kong gateway deployment to be patched

## _KONG_GATEWAY_DEPLOYMENT_NAMESPACE_ - the namespace of the Kong gateway deployment to be patched

## _KONG_GATEWAY_INGRESS_NAME_ - the name of the ingress resource to be patched

## _KONG_GATEWAY_INGRESS_NAMESPACE_ - the namespace of the ingress resource to be patched

## _UPSTREAM_TELEMETRY_HOST_NAME_ - The name of the telemetry service (defaults to apiclarity-apiclaritzy.apiclarity:9000)


Once the plugin is deployed, traces will be sent to APIClarity to start learning specs.
  


