# APIClarity

![APIClarity](API_clarity.svg "APIClarity")

Reconstruct [OpenAPI Specifications](https://spec.openapis.org/oas/latest.html)
from real-time workload traffic seamlessly.

## Microservices API challenges

- Not all applications have an OpenAPI specification available
- How can we get this for legacy or external applications?
- Detect whether microservices still use deprecated APIs (a.k.a. Zombie APIs)
- Detect whether microservices use undocumented APIs (a.k.a. Shadow APIs)
- Generate OpenAPI specifications without code instrumentation or
  modifying existing workloads (seamless documentation)

## Solution

- Capture all API traffic in an existing environment using a service mesh
  framework (e.g. [Istio](https://istio.io/))
- Construct an OpenAPI specification by observing API traffic or upload a
  reference OpenAPI spec
- Review, modify and approve automatically generated OpenAPI specs
- Alert on any differences between the approved API specification and the API
  calls observed at runtime; detects shadow & zombie APIs
- UI dashboard to audit and monitor the findings

## High level architecture

![High level architecture](diagram.jpg "High level architecture")

## Building

### Building UI and backend in docker

```shell
DOCKER_IMAGE=<your repo>/apiclarity DOCKER_TAG=<your tag> make push-docker
# Modify the image name of the APIClarity deployment in ./deployment/apiclarity.yaml
```

### Building UI

```shell
make ui
```

### Building Backend

```shell
make backend
```

## Installation in a K8s cluster

1. Make sure that Istio is installed and running in your cluster.
   See the [Official installation instructions](https://istio.io/latest/docs/setup/getting-started/#install)
   for more information.

2. Add Helm repo

   ```shell
   helm repo add apiclarity https://apiclarity.github.io/apiclarity
   ```

3. Deploy APIClarity with Helm

   ```shell
   helm install --set 'global.namespaces={namespace1,namespace2}' apiclarity apiclarity/apiclarity -n apiclarity
   ```
  **Note**:
  namespace1 and namespace2 are the namespaces where the Envoy Wasm filters will be deployed to allow traffic tracing.

4. Port forward to APIClarity UI:

   ```shell
   kubectl port-forward -n apiclarity svc/apiclarity-apiclarity 9999:8080
   ```

5. Open APIClarity UI in the browser: <http://localhost:9999/>
6. Generate some traffic in the applications in the traced namespaces and check
   the APIClarity UI :)

## Configurations

The file `values.yaml` is used to deploy and configure APIClarity on your cluster via Helm.

1. Set `RESPONSE_HEADERS_TO_IGNORE` and `REQUEST_HEADERS_TO_IGNORE` with a space separated list of headers to ignore when reconstructing the spec.

    Note: Current values defined in `headers-to-ignore-config` ConfigMap

## Testing with a demo application

A good demo application to try APIClarity with is the [Sock Shop Demo](https://microservices-demo.github.io/).

To deploy the Sock Shop Demo follow these steps:

1. Create the `sock-shop` namespace and enable Istio injection:

   ```shell
   kubectl create namespace sock-shop
   kubectl label namespaces sock-shop istio-injection=enabled
   ```

2. Deploy the Sock Shop Demo to your cluster:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/microservices-demo/microservices-demo/master/deploy/kubernetes/complete-demo.yaml
   ```

3. Deploy APIClarity in the `sock-shop` namespace:

   ```shell
   helm install --set 'global.namespaces={sock-shop}' apiclarity apiclarity/apiclarity -n apiclarity
   ```

4. Find the NodePort to access the Sock Shop Demo App

   ```shell
   $ kubectl describe svc front-end -n sock-shop
   [...]
   NodePort:                 <unset>  30001/TCP
   [...]
   ```

   Use this port together with your node IP to access the demo webshop and run
   some transactions to generate data to review on the APIClarity dashboard.

## Running locally with demo data

1. Build UI & backend locally as described above:

   ```shell
   make ui && make backend
   ```

2. Copy the built site:

   ```shell
   cp -r ./ui/build ./site
   ```

3. Run backend and frontend locally using demo data:

   ```shell
   DATABASE_DRIVER=LOCAL FAKE_TRACES=true FAKE_TRACES_PATH=./backend/pkg/test/trace_files \
   ENABLE_DB_INFO_LOGS=true ./backend/bin/backend run
   ```

4. Open APIClarity UI in the browser: <http://localhost:8080/>

## Contributing

Pull requests and bug reports are welcome.

For larger changes please create an Issue in GitHub first to discuss your
proposed changes and possible implications.

## License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
