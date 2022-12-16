#!/bin/bash

APICLARITY_AGENT_IMAGE=apiclarity-bigip-agent

docker build . -t ${APICLARITY_AGENT_IMAGE}