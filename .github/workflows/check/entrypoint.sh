#!/bin/sh

set -e

if [ -z "$(gofmt -l .)" ]; then echo "Format OK"; else echo "Format Fail. Run \"go fmt ./...\""; exit 1; fi
go test -v -cover -tags unit ./... -coverprofile=coverage.txt -covermode=atomic
