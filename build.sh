#!/usr/bin/env bash

CGO_ENABLED=0 GO111MODULE=on GOARCH=amd64 GOOS=windows go build -i -v -ldflags="-X main.CARCASS_VERSION=$(git describe --always --tags --long --dirty)"
