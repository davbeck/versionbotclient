#!/usr/bin/env bash

# this will build the project for both 32 and 64 bit macOS and then combine them into a single universal binary
NAME="versionbotclient"

export GOOS=darwin

export GOARCH=amd64
go build -o "build/$NAME-64"

export GOARCH=386
go build -o "build/$NAME-32"

lipo -create "build/$NAME-64" "build/$NAME-32" -o "build/$NAME"