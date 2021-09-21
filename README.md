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
docker build -t <your repo>/apiclarity .
docker push <your repo>/apiclarity
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
2. Deploy APIClarity in K8s. It will be deployed in a new namespace `apiclarity`:

   ```shell
   kubectl apply -f deployment/apiclarity.yaml
   ```

3. Verify that APIClarity is running:

   ```shell
   $ kubectl get pods -n apiclarity
   NAME                        READY   STATUS    RESTARTS   AGE
   apiclarity-5df5fd6d98-h8v7t   1/1     Running   2          15m
   mysql-6ffc46b7f-bggrv       1/1     Running   0          15m
   ```

4. Initialize and pull the `wasm-filters` submodule:

   ```shell
   git submodule init wasm-filters
   git submodule update wasm-filters
   cd wasm-filters
   ```

5. Deploy the Envoy Wasm filter for capturing the traffic:

   Run the Wasm deployment script for selected namespaces to allow traffic
   tracing.

   Tracing is accomplished by patching the Istio sidecars within the pods
   to load the APIClarity Wasm filter. So ensure [Istio sidecar injection](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/)
   is enabled for all namespaces you intend to trace before deploying anything
   to that namespace.

   The script will automatically:
   - Deploy the Wasm filter binary as a config map
   - Deploy the Istio Envoy filter to use the Wasm binary
   - Patch all deployment annotations within the selected namespaces to mount
     the Wasm binary

   ```shell
   ./deploy.sh <namespace1> <namespace2> ...
   ```

   **Note**:
   To build the Wasm filter from source instead of using the pre-built binary,
   please follow the instructions in the [wasm-filters](https://github.com/apiclarity/wasm-filters)
   repository.
6. Port forward to APIClarity UI:

   ```shell
   kubectl port-forward -n apiclarity svc/apiclarity 9999:8080
   ```

7. Open APIClarity UI in the browser: <http://localhost:9999/>
8. Generate some traffic in the applications in the traced namespaces and check
   the APIClarity UI :)

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

3. From the APIClarity git repository deploy the Wasm filter in the `sock-shop`
   namespace:

   ```shell
   cd apiclarity/wasm-filters
   ./deploy.sh sock-shop
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
   FAKE_TRACES=true FAKE_TRACES_PATH=./backend/pkg/test/trace_files \
   ENABLE_DB_INFO_LOGS=true ./backend/bin/backend run
   ```

4. Open APIClarity UI in the browser: <http://localhost:8080/>

## Contributing

Pull requests and bug reports are welcome.

For larger changes please create an Issue in GitHub first to discuss your
proposed changes and possible implications.

## License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
