#!/bin/bash
set -euo pipefail

env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o my-plugin.wasm main.go
