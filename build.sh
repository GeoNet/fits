#!/bin/bash -eu

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

if [ $# -eq 0 ]; then
  echo Error: please supply a project to build. Usage: ./build.sh project [project]
  exit 1
fi

# code will be compiled in this container
BUILDER_IMAGE='ghcr.io/geonet/base-images/golang:1.23.5-alpine3.21'
RUNNER_IMAGE='ghcr.io/geonet/base-images/static:latest'

VERSION='git-'$(git rev-parse --short HEAD)
ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')

for i in "$@"; do
  if [[ "${i}" == "fits-api" ]]; then
    SOURCEPATH="."
  else
    SOURCEPATH="./dapper"
  fi

  mkdir -p "${SOURCEPATH}/cmd/${i}/assets"
  dockerfile="Dockerfile"
  if test -f "${SOURCEPATH}/cmd/${i}/Dockerfile"; then
    dockerfile="${SOURCEPATH}/cmd/${i}/Dockerfile"
  else
    cat Dockerfile_template > $dockerfile
    echo "CMD [\"/${i}\"]" >> $dockerfile
  fi

  docker build \
    --build-arg=BUILD="$i" \
    --build-arg=RUNNER_IMAGE="$RUNNER_IMAGE" \
    --build-arg=BUILDER_IMAGE="$BUILDER_IMAGE" \
    --build-arg=GIT_COMMIT_SHA="$VERSION" \
    --build-arg=ASSET_DIR="${SOURCEPATH}/cmd/$i/assets" \
    --build-arg=SOURCEPATH="$SOURCEPATH" \
    -t "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION" \
    -f $dockerfile .

  # tag latest.  Makes it easier to test with compose.
  docker tag \
    "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION" \
    "${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest"

done

# vim: set ts=4 sw=4 tw=0 et:
