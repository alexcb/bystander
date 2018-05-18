#!/bin/sh
set -e
export GOPATH=`pwd`
go build -o bystander cmd/bystander/main.go
