#!/bin/bash
# for pushing th image from travis
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
export VER=`grep "const version = " cmd/netperf-operator/main.go | cut -f2 -d'"'` 
export GIT=`git rev-parse --short HEAD` 
echo "pushing version v${VER}-${GIT}"
docker push tailoredcloud/netperf-operator:v${VER}-${GIT} 