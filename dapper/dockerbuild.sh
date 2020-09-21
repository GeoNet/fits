#!/usr/bin/env bash

set -e # Exit if any command fails

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION="git-$(git rev-parse --short HEAD)"

eval $(aws ecr get-login --no-include-email --region ap-southeast-2)

for i in "$@"
do
    rm -f docker-build-temp

    echo "FROM golang:1.12-alpine as build" >> docker-build-temp
    echo "WORKDIR /go/src/github.com/GeoNet/fits/dapper" >> docker-build-temp
    echo "COPY ./cmd/${i} ./cmd/${i}" >> docker-build-temp
    echo "COPY ./dapperlib ./dapperlib" >> docker-build-temp
    echo "COPY ./internal ./internal" >> docker-build-temp
    echo "COPY ./vendor ./vendor" >> docker-build-temp
    echo "RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags \"-X main.Prefix=/usr/local -extldflags -static\" -installsuffix cgo -o ${i} cmd/${i}/*.go">> docker-build-temp
    echo "RUN apk --update add ca-certificates" >> docker-build-temp

    echo "FROM scratch" >> docker-build-temp
    echo "COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt" >> docker-build-temp
    echo "COPY --from=build /go/src/github.com/GeoNet/fits/dapper/${i} ./${i}" >> docker-build-temp
    echo "CMD [\"./${i}\"]" >> docker-build-temp

    docker build -t ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest -f docker-build-temp . # Or we use the project generic Dockerfile (which is passed the cmd as a build arg if the build needs it)

    #TODO: We need some quay.io logic in here

    docker tag ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

done