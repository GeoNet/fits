#!/bin/bash -e

# pushes Docker images for the arg list.
#
# usage: ./push.sh project [project]

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"

do
	docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION
	docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
done
