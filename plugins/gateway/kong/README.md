## APIClarity Kong plugin

### Deploy

run:


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