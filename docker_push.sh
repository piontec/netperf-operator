#!/bin/bash
# for pushing th image from travis
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
VER=`grep "const version = " cmd/netperf-operator/main.go | cut -f2 -d'"'` GIT=`git rev-parse --short HEAD` docker push tailoredcloud/netperf-operator:v${VER}-${GIT} 