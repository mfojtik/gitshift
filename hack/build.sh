#!/bin/bash

set -e

echo "-> Building 'bin/gs' binary ..."
mkdir -p bin
GOOS=linux GOARCH=arm GOARM=7 go build -o ./bin/gs ./main.go

echo -> "Building gshift-base-arm:latest image ..."
docker build -f Dockerfile.base -t docker.io/mfojtik/gshift-base-arm:latest .

echo -> "Pushing gshift-base-arm:latest image ..."
docker push docker.io/mfojtik/gshift-base-arm:latest
