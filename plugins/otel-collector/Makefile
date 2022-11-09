.PHONY: default test
default: lint

../../bin/golangci-lint:
	@(cd ../../ ; $(MAKE) bin/golangci-lint)

# TODO: remove skip when package passes typecheck linter
.PHONY: lint
lint: ../../bin/golangci-lint ## Run linter
	cd apiclarityexporter ; ../../../bin/golangci-lint run \
	  --skip-files '../../../../../../go/pkg/mod/github.com/klauspost/compress'

.PHONY: fix
fix: ../../bin/golangci-lint
	cd apiclarityexporter ; ../../../bin/golangci-lint run --fix

.PHONY: test
test: ## Run Unit Tests
	cd apiclarityexporter ; go test
