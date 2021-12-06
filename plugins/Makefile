# Project variables
DOCKER_REGISTRY ?= ghcr.io/apiclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_TAG ?= ${VERSION}
TYK_VERSION ?= v3.2.2

.PHONY: docker
docker: docker-kong docker-tyk docker-gopacket

.PHONY: docker-kong
docker-kong:
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REGISTRY}/kong-plugin:${DOCKER_TAG} --file Dockerfile.kong .

.PHONY: docker-gopacket
docker-gopacket:
	@DOCKER_BUILDKIT=1 docker build --build-arg VERSION=${VERSION} -t ${DOCKER_REGISTRY}/gopacket-agent:${DOCKER_TAG} --file Dockerfile.gopacket .

.PHONY: docker-tyk
docker-tyk:
	@(cd gateway/tyk/${TYK_VERSION} && \
	go mod edit -replace github.com/apiclarity/apiclarity/plugins/api@v0.0.0=./../api && \
	go mod edit -replace github.com/apiclarity/apiclarity/plugins/common@v0.0.0=./../common && \
	 cd -)
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_REGISTRY}/tyk-plugin-${TYK_VERSION}:${DOCKER_TAG} --file Dockerfile.tyk.${TYK_VERSION} .
	@(cd gateway/tyk/${TYK_VERSION} && \
	go mod edit -replace github.com/apiclarity/apiclarity/plugins/api@v0.0.0=./../../../api && \
	go mod edit -replace github.com/apiclarity/apiclarity/plugins/common@v0.0.0=./../../../common && \
	 cd -)

.PHONY: push-docker
push-docker: push-docker-kong push-docker-tyk push-docker-gopacket

.PHONY: push-docker-kong
push-docker-kong: docker-kong
	@echo "Publishing Kong plugin Docker image ..."
	docker push ${DOCKER_REGISTRY}/kong-plugin:${DOCKER_TAG}

.PHONY: push-docker-tyk
push-docker-tyk: docker-tyk
	@echo "Publishing Tyk plugin Docker image ..."
	docker push ${DOCKER_REGISTRY}/tyk-plugin-${TYK_VERSION}:${DOCKER_TAG}

.PHONY: push-docker-gopacket
push-docker-gopacket: docker-gopacket
	@echo "Publishing gopacket extension Docker image ..."
	docker push ${DOCKER_REGISTRY}/gopacket-agent:${DOCKER_TAG}
