#!/usr/bin/env sh

# Download proto3 from github releases:
#   https://github.com/protocolbuffers/protobuf/releases/tag/v27.1
#
# Install protoc generator for golang from go install:
#   $ go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
#   $ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
# And make sure to insert go bin path to your PATH:
#   $ export PATH="$PATH:$(go env GOPATH)/bin"
#
# See algo:
#   https://grpc.io/docs/languages/go/quickstart/

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    server.proto
