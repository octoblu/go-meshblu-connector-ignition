#!/bin/bash

BUILD_DIR="$HOME/Projects/octoblu/meshblu-connector-test"

build_it() {
  env GOOS="linux" GOARCH="amd64" \
    go build -a -tags netgo -installsuffix cgo -ldflags "-w" \
    -o "${BUILD_DIR}/start" .
}

fatal() {
  local message="$1"
  echo "$message"
  exit 1
}

main() {
  build_it
}

main "$@"