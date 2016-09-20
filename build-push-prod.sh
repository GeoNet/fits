#!/bin/bash -e

# Builds and pushes Docker images for the arg list.  Tags prod
#
# usage: ./build-push-prod.sh project [project]

./build.sh $@

VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"
do

		docker tag 862640294325.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION 862640294325.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:prod
		docker push 862640294325.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:prod

done
