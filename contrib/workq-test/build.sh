#!/bin/sh

GOOS=linux GOARCH=arm GOARM=7 go build -o workq-test ./main.go && \
docker build -t docker.io/mfojtik/work-test-arm:latest . && \
docker push docker.io/mfojtik/work-test-arm:latest
