# API Clarity
Open source for API traffic visibility in K8s clusters

# Microservices API challenges
* Not all applications have their Open API specification available.​
* How can we get this for our legacy or external applications ?​
* Ability to detect that microservices still use deprecated APIs (a.k.a. Zombie APIs)​
* Ability to detect that microservices use undocumented APIs (a.k.a. Shadow APIs)​
* Ability to get Open API specifications without code instrumentation or modifying existing workloads (seamless documentation)

# Solution
* Capture all API traffic in an existing environment using a service-mesh framework​
* Construct the Open API specification by observing the API traffic​
* Allow the User to upload Open API spec, review, modify and approve generated Open API specs​
* Alert the user on any difference between the approved API specification and the one that is observed in runtime, detects shadow & zombie APIs​
* UI dashboard to audit and monitor the API findings

![High level diagram](diagram.jpg "High level diagram")

# Building
## Building UI and backend in docker
```
docker build -t <your repo>/apiclarity .
docker push <your repo>/apiclarity
# Modify the image name of the apiclarity deployment in ./deployment/apiclarity.yaml
```
## Building UI
```
make ui
```

## Building Backend
```
export GO111MODULE=on && make backend
```

# Installation in a K8s cluster
1. Make sure that Istio is installed and running in your cluster:
   
   1.1. Istio that is deployed as part of SecureCN installation.
   
   1.2 Official installation [instructions](https://istio.io/latest/docs/setup/getting-started/#install).
   

2. Deploy APIClarity in K8s (will be deployed in a new namespace named apiclarity):
```
kubectl apply -f deployment/apiclarity.yaml
```
3. Check that APIClarity is running:
```
kubectl get pods -n apiclarity
NAME                        READY   STATUS    RESTARTS   AGE
apiclarity-5df5fd6d98-h8v7t   1/1     Running   2          15m
mysql-6ffc46b7f-bggrv       1/1     Running   0          15m
```
4. Build the Envoy WASM filter for capturing the traffic:
```
git submodule init wasm-filters
git submodule update wasm-filters
cd wasm-filters
make docker_build && ls -l bin 
```
5. Run the WASM deployment script for selected namespaces to allow traffic tracing.
The script will automatically:
   
   - Deploy the WASM filter binary as a config map.
   
   - Deploy the Istio Envoy filter to use the WASM binary.
   
   - Patch all deployment annotations within the selected namespaces to mount the WASM binary.

```
./deploy.sh <namespace1> <namespace2> ...
```
6. Port forward to APIClarity UI:
```
kubectl port-forward -n apiclarity svc/apiclarity 9999:8080
```

7. Open APIClarity UI in the browser: [http://localhost:9999/](http://localhost:9999/)

8. Generate some traffic in the applications (e.g. [sock-shop demo](https://github.com/microservices-demo/microservices-demo)) in the traced namespaces and check APIClarity UI :)


## Running locally with demo data
1. Build UI & backend locally as described above:
```
make ui && make backend
```
2. Copy the built site:
```
cp -r ./ui/build ./site
```
3. Run backend and frontend locally using demo data:
```
FAKE_DATA=true ENABLE_DB_INFO_LOGS=true ./backend/bin/backend run
```
4. Open APIClarity UI in the browser: [http://localhost:8080/](http://localhost:8080/)


