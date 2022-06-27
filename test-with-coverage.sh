#!/usr/bin/env bash
set -e
go test -coverprofile=c.out -v ./...
go tool cover -html=c.out
