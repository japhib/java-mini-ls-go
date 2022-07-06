#!/usr/bin/env bash

# exit early if a step fails
set -e

go test ./...
./lint.sh