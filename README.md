# APIClarity

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://docs.openclarity.io/img/footer-logos/OC_logo_H_1C_white.svg">
  <source media="(prefers-color-scheme: light)" srcset="https://docs.openclarity.io/img/color-logo/logo.svg">
  <img alt="Openclarity logo" src="https://docs.openclarity.io/img/color-logo/logo.svg" width="50%">
</picture>

APIClarity is a modular tool that addresses several aspects of API Security, focusing specifically on [OpenAPI](https://spec.openapis.org/oas/latest.html) based APIs.

APIClarity is the tool responsible for API Security in the [OpenClarity platform](https://openclarity.io).

APIClarity approaches API Security in 2 different ways:
- Captures all API traffic in a given environment and performs a set of security analysis to discover all potential security problems with detected APIs
- Actively tests API endpoints to detect security issues in the implementation of such APIs.

## OpenAPI automatic reconstruction
Both approaches described above are way more effective when APIClarity is primed with the OpenAPI specifications of the APIs analyzed or tested. However, not all applications have an OpenAPI specification available. For this reason one of the main functionality of APIClarity is the automatic reconstruction of OpenAPI specifications based on observed API traffic. In this case, users have the ability to review and approve the reconstructed specifications.

## Security Modules
APIClarity is structured in a modular architecture, which allows to easily add new functionalities.

In the following a brief description of the modules currently implemented:

- **Spec Diffs** This module compares the API traces with the OAPI specifications provided by the user or previously reconstructed. The result of this comparison provides:
    - List of API endpoints that are observed but not documented in the specs, i.e. _Shadow APIs_;
    - List of API endpoints that are observed but marked as deprecated in the specs, i.e. _Zombie APIs_;
    - List of difference between of the APIs observed and their documented specification.
- [**Trace Analyzer**](./backend/pkg/modules/internal/traceanalyzer/README.md) This module analyzes path, headers and body of API requests and responses to discover potential security issues, such as weak authentications, exposure of sensitive information, potential Broken Object Level Authorizations (BOLA) etc.
- [**BFLA Detector**](./backend/pkg/modules/internal/bfla/README.md) This module detects potential Broken Function Level Authorization. In particular it observes the API interactions and build an authorization model that captures what clients are supposed to be authorized to make the various API calls. Based on such authorization model it then signals violations which may represent potential issues in the API authorization procedures.
- **Fuzzer** This module actively tests API endpoints based on their specification attempting in discovering security issues in the API server implementation.

## High level architecture

![High level architecture](diagram.png "High level architecture")

## Supported traffic source integrations

APIClarity supports integrating with the following traffic sources. Install APIClarity and follow the instructions per required integration.

* Istio Service Mesh
  * Make sure that Istio 1.10+ is installed and running in your cluster.
  See the [Official installation instructions](https://istio.io/latest/docs/setup/getting-started/#install)
  for more information.

* Tap via a DaemonSet
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/taper)

* Kong API Gateway
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/kong)

* Tyk API Gateway
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/tyk)

* OpenTelemetry Collector (traces only)
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/otel-collector)

The integrations (plugins) for the supported traffic sources above are located in the [plugins directory within the codebase](https://github.com/openclarity/apiclarity/tree/master/plugins) and implement the [plugins API](https://github.com/openclarity/apiclarity/tree/master/plugins/api) to export the API events to APIClarity. To enable and configure the supported traffic sources, please check the ```trafficSource:``` section in [Helm values](https://github.com/openclarity/apiclarity/blob/master/charts/apiclarity/values.yaml).
Contributions of integrations with additional traffic sources are more than welcome!

## Supported Trace Sources
APIClarity can support with the following trace sources and follow the instructions per required integration.

* Apigee X Gateway
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/apigeex)
* BIG-IP LTM Load balancer
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/f5-bigip)
* Kong
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/kong)
* Tyk
  * [Integration instructions](https://github.com/openclarity/apiclarity/tree/master/plugins/gateway/tyk)

## Getting started

To get started, see the [APIClarity documentation on the OpenClarity site](https://openclarity.io/docs/apiclarity/getting-started/).

## Contributing

Pull requests and bug reports are welcome.

For larger changes please create an Issue in GitHub first to discuss your
proposed changes and possible implications.

## Contributors

https://panoptica.app

## License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
