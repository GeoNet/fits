#!/bin/bash -e
# simplified version of geonet-web's docker.sh script.  Will abort if any command fails (-e on hashbang above).

# code will be compiled in this container
BUILD_CONTAINER=golang:1.6.1-alpine

# use a temp dir so we can freely copy new things into it without clobbering our git workspace
PKGNAME=fits
DOCKER_TMP=docker-build-tmp

# Prefix for the logs
BUILD='-X github.com/GeoNet/fits/vendor/github.com/GeoNet/log/logentries.Prefix=git-'`git rev-parse --short HEAD`

mkdir -p $DOCKER_TMP
chmod +s $DOCKER_TMP
mkdir -p ${DOCKER_TMP}/common/etc/ssl/certs
mkdir -p ${DOCKER_TMP}/common/usr/share

# build the app in the build container, output to a local temp directory (not in git), then use docker build to make a container using 'scratch'
# Assemble common resource for ssl and timezones
docker run -e "GOBIN=/usr/src/go/src/github.com/GeoNet/${PKGNAME}/${DOCKER_TMP}" -e "GOPATH=/usr/src/go" -e "CGO_ENABLED=0" -e "GOOS=linux" -e "BUILD=${BUILD}" --rm \
	-v "${PWD}":/usr/src/go/src/github.com/GeoNet/${PKGNAME} \
	-w /usr/src/go/src/github.com/GeoNet/${PKGNAME} ${BUILD_CONTAINER} \
	go install -a -ldflags "${BUILD}" -installsuffix cgo ./...; \
	cp /etc/ssl/certs/ca-certificates.crt ${DOCKER_TMP}/common/etc/ssl/certs; \
	cp -Ra /usr/share/zoneinfo ${DOCKER_TMP}/common/usr/share

# Assemble common resource for user.
echo "nobody:x:65534:65534:Nobody:/:" > ${DOCKER_TMP}/common/etc/passwd

cp fits.json ${DOCKER_TMP}/
cp -R charts.html css js images ${DOCKER_TMP}/

docker build -t quay.io/geonet/fits:latest .
