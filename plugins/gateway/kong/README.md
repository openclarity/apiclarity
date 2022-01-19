## APIClarity Kong plugin

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

### Preserving Client IP Address
Kong is usually deployed behind a Load Balancer (using a Kubernetes Service of type LoadBalancer).
This can result in a loss of actual Client IP address, as Kong will get the IP address of the Load Balancer
as the Client IP address. 

[This](https://docs.konghq.com/kubernetes-ingress-controller/2.1.x/guides/preserve-client-ip/) guide lays out different methods of solving this problem.
