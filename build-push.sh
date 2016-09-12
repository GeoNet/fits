#!/bin/bash -e

# Builds and pushes Docker images for the arg list.
#
# usage: ./build-push.sh project [project]

./build.sh $@

VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"
do
		docker push quay.io/geonet/${i}:$VERSION 
		docker push quay.io/geonet/${i}:latest
done
