#!/usr/bin/env bash

set -e # Exit if any command fails

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION="git-$(git rev-parse --short HEAD)"

eval $(aws ecr get-login --no-include-email --region ap-southeast-2)

for i in "$@"
do
    rm -rf docker-build-tmp
    mkdir docker-build-tmp

    echo "FROM golang:1.15-alpine as build" >> docker-build-tmp/Dockerfile
    echo "WORKDIR /go/src/github.com/GeoNet/fits/dapper" >> docker-build-tmp/Dockerfile
    echo "COPY ./cmd/${i} ./cmd/${i}" >> docker-build-tmp/Dockerfile
    echo "COPY ./dapperlib ./dapperlib" >> docker-build-tmp/Dockerfile
    echo "COPY ./internal ./internal" >> docker-build-tmp/Dockerfile
    echo "COPY ./vendor ./vendor" >> docker-build-tmp/Dockerfile
    echo "RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags \"-X main.Prefix=/usr/local -extldflags -static\" -installsuffix cgo -o ${i} cmd/${i}/*.go">> docker-build-tmp/Dockerfile
    echo "RUN apk --update add ca-certificates" >> docker-build-tmp/Dockerfile

    echo "FROM scratch" >> docker-build-tmp/Dockerfile
    echo "COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt" >> docker-build-tmp/Dockerfile
    echo "COPY --from=build /go/src/github.com/GeoNet/fits/dapper/${i} ./${i}" >> docker-build-tmp/Dockerfile
    if [[ "${i}" == "dapper-api" ]]
    then
        mkdir -p docker-build-tmp/assets
        rsync --archive --quiet --ignore-missing-args cmd/${i}/assets docker-build-tmp/
        echo "COPY ./cmd/dapper-api/assets ./assets/"  >> docker-build-tmp/Dockerfile
    fi 
    
    echo "CMD [\"./${i}\"]" >> docker-build-tmp/Dockerfile

    docker build -t ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest -f docker-build-tmp/Dockerfile .

    #TODO: We need some quay.io logic in here

    docker tag ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
    docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:${VERSION}

done