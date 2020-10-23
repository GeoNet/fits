#!/bin/bash -e

# Builds and pushes Docker images for the arg list.
#
# usage: ./build-push.sh project [project]

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"

do
  if [[ "${i}" == "fits-api" ]]; then
    ./build.sh ${i}
	docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION
	docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
  else
    cd dapper
    ./dockerbuild.sh ${i}
    cd ..
  fi
done
