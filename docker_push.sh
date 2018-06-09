#!/bin/bash
# for pushing th image from travis
export VER=`grep "const version = " cmd/netperf-operator/main.go | cut -f2 -d'"'` 
if [[ $VER = *"-dev"* ]]; then
    echo "This version is marked as dev, not pushing to docker registry"
else
    export GIT=`git rev-parse --short HEAD`
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
    echo "pushing version v${VER}-${GIT}"
    docker push tailoredcloud/netperf-operator:v${VER}-${GIT} 
fi