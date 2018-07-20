#!/bin/sh
set -e

rev=`git rev-parse HEAD`
d=`date +%s`
echo "package bystander\n\nvar gitHash = \"$rev\"\nvar buildTime = $d" > src/bystander/version_generated.go

export GOPATH=`pwd`
go build -o bystander cmd/bystander/main.go
