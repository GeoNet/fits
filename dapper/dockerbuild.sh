#!/usr/bin/env bash

set -e # Exit if any command fails

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION="git-$(git rev-parse --short HEAD)"

eval $(aws ecr get-login --no-include-email --region ap-southeast-2)

for i in "$@"
do

    if test -f "cmd/${i}/Dockerfile"; then
        docker build -t ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest -f cmd/${i}/Dockerfile . # Either we build the image using a specific Dockerfile (defined with project root as the context, despite the Dockerfile being in cmd/${i}
    else
        docker build -t ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest --build-arg COMMAND=${i} . # Or we use the project generic Dockerfile (which is passed the cmd as a build arg if the build needs it)
    fi

    #TODO: We need some quay.io logic in here

    docker tag ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

done