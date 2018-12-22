#!/usr/bin/env bash

_BIN_NAME="bibuild"

[ -d "./pkg" ] || mkdir pkg
rm ./pkg/*
GOOS=linux GOARCH=amd64 go build -ldflags="-w" -o ./pkg/${_BIN_NAME}_linux_amd64
GOOS=windows GOARCH=amd64 go build -ldflags="-w" -o ./pkg/${_BIN_NAME}_windows_amd64.exe
GOOS=darwin GOARCH=amd64 go build -ldflags="-w" -o ./pkg/${_BIN_NAME}_darwin_amd64
