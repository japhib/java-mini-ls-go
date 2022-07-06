#!/usr/bin/env bash

# exit early if a step fails
set -e

./test.sh
./lint.sh