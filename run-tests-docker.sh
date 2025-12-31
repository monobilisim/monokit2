#!/usr/bin/env bash

# Build and run tests inside a Docker container
# Return exit code
docker build -t tests . && docker run --rm tests
