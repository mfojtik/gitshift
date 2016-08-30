#!/bin/bash

set -e
GOOS=linux GOARCH=arm GOARM=7 go build -o gs ./main.go

docker build -f Dockerfile.base -t docker.io/mfojtik/gshift-base-arm:latest .
docker push docker.io/mfojtik/gshift-base-arm:latest

# events
docker build -f Dockerfile.fetch-events -t docker.io/mfojtik/gshift-fetch-events-arm:latest .
docker push docker.io/mfojtik/gshift-fetch-events-arm:latest

docker build -f Dockerfile.process-events -t docker.io/mfojtik/gshift-process-events-arm:latest .
docker push docker.io/mfojtik/gshift-process-events-arm:latest

# comments
docker build -f Dockerfile.fetch-comments -t docker.io/mfojtik/gshift-fetch-comments-arm:latest .
docker push docker.io/mfojtik/gshift-fetch-comments-arm:latest

docker build -f Dockerfile.process-comments -t docker.io/mfojtik/gshift-process-comments-arm:latest .
docker push docker.io/mfojtik/gshift-process-comments-arm:latest

# frontend
docker build -f Dockerfile.frontend -t docker.io/mfojtik/gshift-frontend-arm:latest .
docker push docker.io/mfojtik/gshift-frontend-arm:latest
