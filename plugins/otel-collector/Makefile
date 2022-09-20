COMPONENTS = apiclarityexporter
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_TAG ?= ${VERSION}
OTEL_COLLECTOR_VERSION ?= 0.60.0
APICLARITY_PLUGINS_API_VERSION ?= "v0.0.0-20220915093602-8a11adcdb9e1"

.PHONY: default docker
default: lint
docker: docker-apiclarityexporter

../../bin/golangci-lint:
	@(cd ../../ ; $(MAKE) bin/golangci-lint)

.PHONY: lint
lint: ../../bin/golangci-lint ## Run linter
	for component in $(COMPONENTS) ; do ( cd $$component ; ../../../bin/golangci-lint run ; ) ; done

.PHONY: fix
fix: ../../bin/golangci-lint
	for component in $(COMPONENTS) ; do ( cd $$component ; ../../../bin/golangci-lint run --fix ; ) ; done

.PHONY: test
test: ## Run Unit Tests
	for component in $(COMPONENTS) ; do ( cd $$component ; go test ; ) ; done

.PHONY: otelcol
otelcol: ./otelcol-build/otelcol
./otelcol-build/otelcol: ./otelcol-build/otelcol-builder-${OTEL_COLLECTOR_VERSION}.yaml
	ocb version
	cd ./otelcol-build && ocb --config ./otelcol-builder-${OTEL_COLLECTOR_VERSION}.yaml

./otelcol-build/otelcol-builder-${OTEL_COLLECTOR_VERSION}.yaml: scripts/modify_otelcol_builder.awk
	curl "https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector-releases/v${OTEL_COLLECTOR_VERSION}/distributions/otelcol/manifest.yaml" | \
	awk -v "APICLARITY_PLUGINS_API_VERSION=${APICLARITY_PLUGINS_API_VERSION}" -f scripts/modify_otelcol_builder.awk >$@

# Note: we build the otel-collector in the docker image in order to not need cross-compilation
.PHONY: docker-apiclarityexporter
docker-apiclarityexporter: Dockerfile ./otelcol-build/otelcol-builder-${OTEL_COLLECTOR_VERSION}.yaml
	@DOCKER_BUILDKIT=1 docker build --build-arg OTEL_COLLECTOR_VERSION=${OTEL_COLLECTOR_VERSION} -t otel-apiclarityexporter:${DOCKER_TAG} -t otel-apiclarityexporter:latest .
