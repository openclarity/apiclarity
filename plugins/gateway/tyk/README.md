## APIClarity Tyk gateway plugin

### Prerequisite

* Make sure that Tyk v3.2.2 is running in your K8s cluster.

### Installation using a pre-built image

1. Choose one of the following installation techniques

   1. Script installation
      * Run the following script to add the plugin to your Tyk deployment.
      * The script will add an init container with the plugin image and will mount the plugin into the plugin to /plugins/tyk-plugin.so. 
      * Set the TRACE_SOURCE_TOKEN
        ```shell
        TRACE_SOURCE_TOKEN=$(curl --http1.1 --insecure -s -H 'Content-Type: application/json' -d '{"name":"tyk-plugin","type":"TYK"}' https://localhost:8443/api/control/traceSources|jq -r '.token')
        ```

      * Please set TYK_PROXY_CONTAINER_NAME, TYK_GATEWAY_DEPLOYMENT_NAME and TYK_GATEWAY_DEPLOYMENT_NAMESPACE accordingly:

       ```shell
       TYK_PROXY_CONTAINER_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> ./deploy/deploy.sh
       ```

   2. Helm installation
      * Save APIClarity default chart values
       ```shell
       helm show values apiclarity/apiclarity > values.yaml
       ```
      * Update the values in `trafficSource.tyk`
      * Deploy or Upgrade APIClarity
      ```shell
      helm upgrade --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity --install
      ```
      * A post install job will execute the installation script in your cluster

2. Update your Tyk API definition with the plugin configuration:
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

3. Hot reload Tyk:
```shell
curl -s -H "x-tyk-authorization: $TYK_ADMIN_KEY" http://$TYK_ADMIN_ADDRESS/tyk/reload
```

Refer to the documentation for more details:
https://tyk.io/docs/plugins/supported-languages/golang/#loading-the-plugin

### Building from source

Note: The Tyk plugin has to be compiled with the same version of the Tyk gateway.
This is due to limitation of how the go plugin is being build.
Currently, the plugin is compatible only for Tyk gateway version v3.2.2

If you would like to build the plugin for other versions, you need to add the appropriate go.mod dependencies, and then run (from [plugins directory](https://github.com/openclarity/apiclarity/tree/master/plugins)):

```shell
export TYK_VERSION=<your version>
make docker-tyk
```

Then, push the built image and change the image name in [deploy/patch-deployment.yaml](https://github.com/openclarity/apiclarity/blob/master/plugins/gateway/tyk/deploy/patch-deployment.yaml) accordingly.
Then, run the [./deploy/deploy.sh](https://github.com/openclarity/apiclarity/blob/master/plugins/gateway/tyk/deploy/deploy.sh) script.

### Preserving Client IP Address

Tyk is usually deployed behind a Load Balancer (using a Kubernetes Service of type LoadBalancer).
This can result in a loss of actual Client IP address, as Tyk will get the IP address of the Load Balancer
as the Client IP address.

[This](https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip) guide lays out different methods of solving this problem.

### Instructions to uninstall the Tyk plugin 

Run the following script to remove the plugin from your Tyk deployment. The script will rollback to Tyk deployment previous pre-plugin state and remove all the supporting resources created during the deployment. 

  Please set TYK_PROXY_CONTAINER_NAME, TYK_GATEWAY_DEPLOYMENT_NAME and TYK_GATEWAY_DEPLOYMENT_NAMESPACE accordingly:
    
  ```shell
    TYK_PROXY_CONTAINER_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAME=<name> TYK_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> ./deploy/undeploy.sh
  ```
