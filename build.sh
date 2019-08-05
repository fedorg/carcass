#!/usr/bin/env bash

GOOS=windows go build -i -v -ldflags="-X main.CARCASS_VERSION=$(git describe --always --tags --long --dirty)"
