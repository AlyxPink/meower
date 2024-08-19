#!/bin/sh

echo "Generating protobuf files..."

protoc --proto_path=api/protobuf \
       --go_out=api/gen/ \
       --go_opt=paths=source_relative \
       --go-grpc_out=api/gen \
       --go-grpc_opt=paths=source_relative \
       $(find api/protobuf -name '*.proto' -type f)

echo "Generated protobuf files successfully!"
