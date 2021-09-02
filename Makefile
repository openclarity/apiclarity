e_Y=\033[1;33m
C_C=\033[0;36m
C_M=\033[0;35m
C_R=\033[0;41m
C_N=\033[0m
SHELL=/bin/bash

DOCKER_REGISTRY ?= gcr.io/eticloud/k8sec
REPO ?= apiclarity
VERSION ?= $(shell git rev-parse HEAD)
IMAGE_NAME ?= $(DOCKER_REGISTRY)/$(REPO):$(VERSION)

# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help ui backend api docker push-docker

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

ui: ## Build UI
	@(echo "Building UI ..." )
	@(cd ui; npm i ; npm run build; )
	@ls -l ui/build  

backend: ## Build Backend
	@(echo "Building Backend ..." )
	@(cd backend && go build -o bin/backend cmd/backend/main.go && ls -l bin/)

backend_linux: ## Build Backend Linux
	@(echo "Building Backend linux..." )
	@(cd backend && GOOS=linux go build -o bin/backend_linux cmd/backend/main.go && ls -l bin/)

backend_test: ## Build Backend test
	@(echo "Building Backend test ..." )
	@(cd backend && go build -o bin/backend_test cmd/test/main.go && ls -l bin/)

api: ## Generating API code
	@(echo "Generating API code ..." )
	@(cd api; ./generate.sh)

docker: ## Build Docker image 
	@(echo "Building docker image ..." )
	docker build -t $(IMAGE_NAME) .

push-docker: docker ## Build and Push Docker image
	@echo "Publishing Docker image ..."
	docker push $(IMAGE_NAME)

test: ## Run Unit Tests
	@(cd backend && FAKE_DATA=true go test ./pkg/...)

clean: clean-ui clean-backend ## Clean all build artifacts

clean-ui: 
	@(rm -rf ui/build ; echo "UI cleanup done" )

clean-backend: 
	@(rm -rf bin ; echo "Backend cleanup done" )
