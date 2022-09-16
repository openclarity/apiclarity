
COMPONENTS = exporter

.PHONY: default
default: lint

../../bin/golangci-lint:
	@(cd ../../ ; $(MAKE) bin/golangci-lint)

.PHONY: lint
lint: ../../bin/golangci-lint ## Run linter
	for component in $(COMPONENTS) ; do ( cd pkg/$$component ; ../../../../bin/golangci-lint run ; ) ; done

.PHONY: test
test: ## Run Unit Tests
	for component in $(COMPONENTS) ; do ( cd pkg/$$component ; go test ; ) ; done

./otelcol-dev: ./otelcol-builder.yaml
	ocb version
	ocb --config ./otelcol-builder.yaml --output-path ./otelcol-build