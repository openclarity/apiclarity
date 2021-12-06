## APIClarity Gopacket Traffic Source

### Deploy

run:


```shell
UPSTREAM_TELEMETRY_HOST_NAME=<address> NAMESPACES_TO_TAP=<namespaces> ./deploy/deploy.sh
```

Where:

UPSTREAM_TELEMETRY_HOST_NAME - The address of the telemetry service (defaults to apiclarity-apiclarity.apiclarity:9000)

NAMESPACES_TO_TAP - list of namespaces to tap in the format: "ns1 ns2 ns3" (defaults to default namespace)
