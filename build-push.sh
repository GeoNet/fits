#!/bin/bash -e

# Builds and pushes Docker images for the arg list.
#
# usage: ./build-push.sh project [project]

./build.sh $@

ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
VERSION='git-'`git rev-parse --short HEAD`

for i in "$@"
do
		docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:$VERSION 
		docker push ${ACCOUNT}.dkr.ecr.ap-southeast-2.amazonaws.com/${i}:latest
done

#cd dapper
#./dockerbuild.sh dapper-api dapper-db-archive dapper-db-ingest dapper-db-meta-ingest
#cd ..
