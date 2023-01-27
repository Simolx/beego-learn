#!/bin/bash

set -x
echo "start to build images"
cp -r ../go.mod ../go.sum ../cmd ../conf .
docker build -t zkserver .
rm -rf ./go.mod ./go.sum ./cmd ./conf
echo "build succeed"
set +x

