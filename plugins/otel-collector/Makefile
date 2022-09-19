
COMPONENTS = apiclarityexporter
DOCKER_REGISTRY ?= ghcr.io/openclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_TAG ?= ${VERSION}

.PHONY: default
default: lint

../../bin/golangci-lint:
	@(cd ../../ ; $(MAKE) bin/golangci-lint)

.PHONY: lint
lint: ../../bin/golangci-lint ## Run linter
	for component in $(COMPONENTS) ; do ( cd $$component ; ../../../bin/golangci-lint run ; ) ; done

.PHONY: test
test: ## Run Unit Tests
	for component in $(COMPONENTS) ; do ( cd $$component ; go test ; ) ; done

.PHONY: otelcol
otelcol: ./otelcol-build/otelcol
./otelcol-build/otelcol: ./otelcol-build/otelcol-builder.yaml
	ocb version
	ocb --config ./otelcol-build/otelcol-builder.yaml --output-path ./otelcol-build

.PHONY: docker-apiclarityexporter
docker-apiclarityexporter: Dockerfile
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REGISTRY}/otel-apiclarityexporter:${DOCKER_TAG} .
