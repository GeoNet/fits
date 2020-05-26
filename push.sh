#!/usr/bin/env bash

set -eu # Exit if any command fails

for i in "$@"
do
    docker push "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest"
    docker push "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}"
done
