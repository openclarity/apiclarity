## APIClarity Kong plugin

### Deploy

run:


```shell
KONG_PROXY_CONTAINER_NAME=<name> KONG_GATEWAY_DEPLOYMENT_NAME=<name> KONG_GATEWAY_INGRESS_NAME=<name> ./deploy
```

Where:

KONG_PROXY_CONTAINER_NAME - the name of the proxy container in Kong gateway (default to proxy)

KONG_GATEWAY_DEPLOYMENT_NAME - the name of the Kong gateway deployment to be patched

KONG_GATEWAY_INGRESS_NAME - the name of the ingress resource to be patched

Once the plugin is deployed, traces will be sent to APIClarity to start learning specs.