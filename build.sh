#!/bin/bash -e

# Builds Docker images for the arg list.  These must be project directories
# where this script is executed.
#
# Builds a statically linked executable and adds it to the container.
# Adds the assets dir from each project to the container e.g., origin/assets
# It is not an error for the assets dir to not exist.
# Any assets needed by the application should be read from the assets dir
# relative to the executable. 
#
# usage: ./build.sh project [project]

: "${VERSION="$(git rev-parse --short HEAD)"}"
ecode=0
if [[ -z ${ACCOUNT+x} ]]; then
    ecode=255
    echo "\$ACCOUNT env var not set, this is only okay for a localbuild!"
    ACCOUNT='LOCAL'
fi

set -euo pipefail

if [ $# -eq 0 ]; then
    echo Error: please supply a project to build. Usage: ./build.sh project [project]
    exit 1
fi

for i in "$@"
do
    docker build --build-arg=BUILD="$i" \
        --build-arg=ASSET_DIR="./cmd/$i/assets" \
        --build-arg=GIT_COMMIT_SHA="$VERSION" \
        -t "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION" \
        -f Dockerfile .

    # Tag the built container as :latest
    docker tag "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION" "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest"

done

exit $ecode

