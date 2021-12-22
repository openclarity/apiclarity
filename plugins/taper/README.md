## APIClarity Passive Tapper Traffic Source

### Installation using a pre-built image

#### Helm installation
* Save APIClarity default chart values
```shell
helm show values apiclarity/apiclarity > values.yaml
```
* Update the values in `trafficSource.tap`
* Deploy or Upgrade APIClarity
```shell
helm upgrade --values values.yaml --create-namespace apiclarity apiclarity/apiclarity -n apiclarity --install
```
* An APIClarity Tap DaemonSet will we deployed to your cluster and will tap the provided namesapces.
