#!/usr/bin/env bash

set -e

# build the builder docker image
docker build -t builder-ecr-get-credentials .

# run builder docker image to build the 'run' image
# share the docker socket, executable and build key
docker run \
    -v /var/run/docker.sock:/var/run/docker.sock \
    builder-ecr-get-credentials
