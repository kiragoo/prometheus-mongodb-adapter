#!/usr/bin/env bash
set -e

docker build -t prometheus-mongodb-adapter .

docker tag prometheus-mongodb-adapter zcloudws/prometheus-mongodb-adapter:latest
docker push zcloudws/prometheus-mongodb-adapter:latest