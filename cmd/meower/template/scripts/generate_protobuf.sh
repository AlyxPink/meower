#!/bin/sh

echo "Generating protobuf files..."

protoc --proto_path=api/proto \
       --go_out=api/proto/ \
       --go_opt=paths=source_relative \
       --go-grpc_out=api/proto/ \
       --go-grpc_opt=paths=source_relative \
       $(find api/proto -name '*.proto' -type f)

echo "Generated protobuf files successfully!"
