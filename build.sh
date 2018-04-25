#!/usr/bin/env bash

export GOPATH="$PWD/../../"
CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -v