#!/bin/sh
set -e

rev=`git rev-parse HEAD`
d=`date +%s`
echo "package bystander\n\nvar gitHash = \"$rev\"\nvar buildTime = $d" > src/bystander/version_generated.go

cd src/bystander
dep ensure
cd ../..

export GOPATH=`pwd`
go test bystander
go build -o bystander cmd/bystander/main.go
