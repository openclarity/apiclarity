## APIClarity Kong plugin

### Prerequisite

* APIClarity backend is running.
* Kong gateway is running in your K8s cluster, and has an Ingress gateway resource.

If you just want to try it out with a demo application, and you don't have kong installed, you can follow these few short steps that will help you to quickly setup a running environment:

1. Deploy sock-shop app:
    ```shell
       kubectl create namespace sock-shop
       
       kubectl apply -f https://raw.githubusercontent.com/microservices-demo/microservices-demo/master/deploy/kubernetes/complete-demo.yaml
    ```
2. Deploy Kong:
    - Using helm3:
    ```shell
       helm repo add kong https://charts.konghq.com
       helm repo update
       helm -n kong install kong/kong --create-namespace --generate-name --set ingressController.installCRDs=false
    ```
    - Using kubectl:
    ```shell
       kubectl apply -f https://bit.ly/kong-ingress-dbless
    ```
3. Wait for pod to be ready:
    ```shell
       watch kubectl get pods -n kong
    ```
4. Configure Ingress:
    ```shell
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
    ```
5. Deploy APIClarity Kong Plugin:
    ```shell
       cd ~/go/src/github.com/openclarity/apiclarity/plugins/gateway/kong
       
       KONG_GATEWAY_DEPLOYMENT_NAME=ingress-kong \
       KONG_GATEWAY_DEPLOYMENT_NAMESPACE=kong \
       KONG_GATEWAY_INGRESS_NAME=catalogue \
       KONG_GATEWAY_INGRESS_NAMESPACE=sock-shop \
       UPSTREAM_TELEMETRY_HOST_NAME=apiclarity-apiclarity.apiclarity:9000 \
       deploy/deploy.sh
    ```
    * Note: If you installed Kong using helm, the deployment name might be different. Please change the KONG_GATEWAY_DEPLOYMENT_NAME env var accordingly.
6. Get LoadBalancer Address:
    ```shell
       export KONG_ADDRESS=$(kubectl get -o jsonpath="{.status.loadBalancer.ingress[0].ip}" service -n kong kong-proxy)
    ```
    Or on EKS get hostname:
    ```shell
       export KONG_ADDRESS=$(kubectl get -o jsonpath="{.status.loadBalancer.ingress[0].hostname}" service -n kong kong-proxy)
    ```    
    * Note:
        - If you installed Kong using helm, the service name might be different.  
7. Run Traffic:
    ```shell
       curl -H 'content-type: application/json' -H 'accept: application/json;charset=UTF-8' $KONG_ADDRESS/catalogue
       curl -H 'content-type: application/json' -H 'accept: application/json;charset=UTF-8' $KONG_ADDRESS/catalogue/size
       curl -H 'content-type: application/json' -H 'accept: application/json;charset=UTF-8' $KONG_ADDRESS/tags
    ```
8. Cleanup:
    
    1. Delete kong installation:
        ```shell
           helm -n kong uninstall kong 
        ```   
          Or if not installed with helm:
        ```shell
            kubectl delete -f https://bit.ly/kong-ingress-dbless 
        ```
    2. Delete sock-shop:
        ```shell
            kubectl delete ns sock-shop 
        ```
### Installation using a pre-built image

1. Choose one of the following installation techniques

    1. Script installation
        * Run the following script to add the plugin to your Kong deployment and Ingress configuration.
        * Please set all env variables accordingly:

        ```shell
       KONG_PROXY_CONTAINER_NAME=<name> \
       KONG_GATEWAY_DEPLOYMENT_NAME=<name> \
       KONG_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> \
       KONG_GATEWAY_INGRESS_NAME=<name> \
       KONG_GATEWAY_INGRESS_NAMESPACE=<namespace> \
       UPSTREAM_TELEMETRY_HOST_NAME=<telemetry service address> ./deploy/deploy.sh
        ```

    2. Helm installation
        * Save APIClarity default chart values
        ```shell
        helm show values apiclarity/apiclarity > values.yaml
        ```
        * Update the values in `trafficSource.kong`
        * Deploy or Upgrade APIClarity
       ```shell
       helm upgrade --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity --install
       ```
        * A post install job will execute the installation script in your cluster

### Building from source

If you would like to customize the plugin, you should run this from the plugins directory:

```shell
make docker-kong
```

You can then push the plugin image to any public registry, and update the helm values to use that image.

### Preserving Client IP Address

Kong is usually deployed behind a Load Balancer (using a Kubernetes Service of type LoadBalancer).
This can result in a loss of actual Client IP address, as Kong will get the IP address of the Load Balancer
as the Client IP address.

[This](https://docs.konghq.com/kubernetes-ingress-controller/2.1.x/guides/preserve-client-ip/) guide lays out different methods of solving this problem.

### Instructions to uninstall the Kong plugin 

Run the following script to remove the plugin from your Kong deployment. The script will rollback to Kong deployment previous pre-plugin state and remove all the supporting resources created during the deployment. 

  Please set all env variables accordingly:
    
  ```shell
       KONG_PROXY_CONTAINER_NAME=<name> \
       KONG_GATEWAY_DEPLOYMENT_NAME=<name> \
       KONG_GATEWAY_DEPLOYMENT_NAMESPACE=<namespace> \
       KONG_GATEWAY_INGRESS_NAME=<name> \
       KONG_GATEWAY_INGRESS_NAMESPACE=<namespace> ./deploy/undeploy.sh
  ```