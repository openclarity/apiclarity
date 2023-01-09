# APIClarity Kuma plugin

This chart installs [APIClarity WASM filter](https://github.com/openclarity/wasm-filters) inside the Kuma dataplane.
This is done by creating a [Kuma Proxy Template](https://kuma.io/docs/2.0.x/policies/proxy-template/).

## Prerequisites

- APIClarity already installed in your Kubernetes cluster

## Configuration

The configuration is divided in two parts, one is for Kuma the other is for APIClarity.

### Kuma

- `kuma.kumaMesh`: Name of the Kuma mesh of the dataplanes where you want to
  install the WASM filter.
- `kuma.kumaService`: Name of the Kuma service (as defined by the
  'kuma.io/service' label) where you want to install the WASM filter.

### APIClarity

- `apiclarity.hostname`: Hostname where APIClarity is installed.
- `apiclarity.port`: TCP port where APIClarity is listenning.
- `apiclarity.plugin.config`: WASM filter configuration.

- `apiclarity.plugin.sha256`: SHA256 sum of the compiled WASM filter (most of
  the time, no need to change it).
- `apiclarity.plugin.wasmFilterURI`: Is the full URI of the compiled WASM filter
  (most of the time, no need to change it).

## Installation

1. Add Helm repository (should already be done as part of APIClarity installation)

    ```shell
    helm repo add apiclarity https://openclarity.github.io/apiclarity
    ```

2. Save the default chart values

    ```shell
    helm show values apiclarity/kuma-plugin > values.yaml
    ```

3. Update `values.yaml` according to your preferences

4. Deploy the plugin with Helm

    ```shell
    helm install --values values.yaml my-apiclarity-kuma-plugin apiclarity/kuma-plugin
    ```

5. Check if the Kuma Proxy Template has been created:

    ```shell
    $ kubectl get proxytemplates.kuma.io
    NAME                      AGE
    apiclarity-kuma-default   6s
    ```
