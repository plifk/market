#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Static analysis scripts
cd $(dirname $0)/..

source scripts/lib.sh
ensure_go_binary golang.org/x/lint/golint
ensure_go_binary honnef.co/go/tools/cmd/staticcheck
ensure_go_binary github.com/securego/gosec/cmd/gosec
ensure_go_binary github.com/client9/misspell/cmd/misspell

echo "Running golint."
test -z "$(golint `go list ./...` | tee /dev/stderr)"

echo "Running go vet."
go vet -all ./...

echo "Running staticcheck toolset."
staticcheck ./...

echo "Running tosec to check possible security issues."
gosec -quiet -exclude G104 ./... # Ignoring gosec unhandled errors warning due to many false positives.

echo "Running mispell."
misspell cmd/**/*.{go,sh} internal/**/*.{go} README.md

echo "Running unparam."
unparam ./...

echo "Running tests with data race detector"
go test -race ./...
