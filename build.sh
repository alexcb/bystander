#!/bin/sh
set -e
set -x

rev=`git rev-parse HEAD`
d=`date +%s`
echo "package bystander\n\nvar gitHash = \"$rev\"\nvar buildTime = $d" > src/bystander/version_generated.go

cd src
unset GOPATH
go test ./...
go build -o ../bystander cmd/bystander/main.go
