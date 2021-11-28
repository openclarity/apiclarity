## APIClarity Tyk gateway plugin

### Installation using pre-built image

1. Make sure that Tyk v3.2.2 is running in your K8s cluster.
2. Run the following script to add the plugin to your Tyk deployment. The script will add an init container with the plugin image and will mount the plugin into the plugin to /plugins/tyk-plugin.so. Please set TYK_PROXY_CONTAINER_NAME, TYK_GATEWAY_DEPLOYMENT_NAME and TYK_GATEWAY_DEPLOYMENT_NAMESPACE accordingly:

```shell
TYK_PROXY_CONTAINER_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> ./deploy/deploy.sh
```

3. Update your Tyk API definition with the plugin configuration:
```shell
curl -s -H "x-tyk-authorization: $TYK_ADMIN_KEY" http://$TYK_ADMIN_ADDRESS/tyk/apis/ -X POST -d '{

    "name": "$MY_API_NAME",
    "custom_middleware": {
       "pre": [],
       "post": [
         {
           "name": "PostGetAPIDefinition",
           "path": "/plugins/tyk-plugin.so"
         }
       ],
      "post_key_auth": [],
      "auth_check": {},
       "response": [
         {
           "name": "ResponseSendTelemetry",
           "path": "/plugins/tyk-plugin.so"
         }
       ],
       "driver": "goplugin"
    }
}'

```
4. Hot reload Tyk:
```shell
curl -s -H "x-tyk-authorization: $TYK_ADMIN_KEY" http://$TYK_ADMIN_ADDRESS/tyk/reload
```

Refer to the documentation for more details:
https://tyk.io/docs/plugins/supported-languages/golang/#loading-the-plugin

### Building from source

Note: The tyk plugin has to be compiled with the same version of the tyk gateway.
This is due to limitation of how the go plugin is being build.
Currently, the plugin is competible only for tyk gateway version v3.2.2

If you would like to build the plugin for other versions, you need to add the appropriate go.mod dependencies, and then run (from [plugins directory](https://github.com/apiclarity/apiclarity/edit/tyk-plugin/plugins/)):
```shell
export TYK_VERSION=<your version>
make docker-tyk
```

Then, push the built image and change the image name in deploy/patch-deployment.yaml accordingly.
Then, run the [./deploy/deploy.sh](https://github.com/apiclarity/apiclarity/blob/tyk-plugin/plugins/gateway/tyk/deploy/deploy.sh) script.
