#!/usr/bin/env bash

set -eu # Exit if any command fails

echo $ACCOUNT
echo $VERSION
for i in "$@"; do
    echo DUMMY: push $i
done
exit 0

for i in "$@"
do
    docker push "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest"
    docker push "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}"
done
