#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

# Static analysis scripts
cd $(dirname $0)/..

source scripts/lib.sh
ensure_go_binary honnef.co/go/tools/cmd/staticcheck
ensure_go_binary github.com/securego/gosec/cmd/gosec
ensure_go_binary github.com/client9/misspell/cmd/misspell
ensure_go_binary mvdan.cc/unparam

source scripts/ci-lint-fmt.sh
go vet -all ./...
staticcheck ./...
gosec -quiet -exclude G104 ./... # Ignoring gosec unhandled errors warning due to many false positives.
misspell cmd/**/*.{go,sh} internal/**/*.{go} README.md
unparam ./...
