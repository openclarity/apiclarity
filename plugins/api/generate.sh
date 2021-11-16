#!/bin/sh

set -euo pipefail

alias goswagger="docker run --rm -it --user $(id -u):$(id -g) -e GOPATH=$GOPATH:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger:v0.27.0"

rm -rf client/*
goswagger generate client -f swagger.yaml -t ./client

cp server/restapi/configure_api_clarity_plugins_telemetries_api.go /tmp/configure_api_clarity_plugins_telemetries_api.go
rm -rf server/*
goswagger generate server -f swagger.yaml -t ./server
cp /tmp/configure_api_clarity_plugins_telemetries_api.go server/restapi/configure_api_clarity_plugins_telemetries_api.go