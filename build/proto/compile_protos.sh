#!/bin/sh

go get -u github.com/golang/protobuf/protoc-gen-go

rm -rf build/proto/_output/atomix-api
rm -rf proto
git clone --branch master https://github.com/atomix/atomix-api.git build/proto/_output/atomix-api
mv build/proto/_output/atomix-api/proto proto

proto_imports="./proto:${GOPATH}/src/github.com/google/protobuf/src:${GOPATH}/src"

protoc -I=$proto_imports --go_out=import_path=atomix/primitive,plugins=grpc:proto proto/atomix/primitive/*.proto
protoc -I=$proto_imports --go_out=Matomix/primitive/primitive.proto=github.com/atomix/atomix-go-node/proto/atomix/primitive,import_path=atomix/headers,plugins=grpc:proto proto/atomix/headers/*.proto
protoc -I=$proto_imports --go_out=import_path=atomix/controller,plugins=grpc:proto proto/atomix/controller/*.proto
protoc -I=$proto_imports --go_out=Matomix/partition/partition.proto=github.com/atomix/atomix-go-node/proto/atomix/partition,import_path=atomix/controller,plugins=grpc:proto proto/atomix/controller/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/counter,plugins=grpc:proto proto/atomix/counter/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/election,plugins=grpc:proto proto/atomix/election/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/list,plugins=grpc:proto proto/atomix/list/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/lock,plugins=grpc:proto proto/atomix/lock/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/log,plugins=grpc:proto proto/atomix/log/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/map,plugins=grpc:proto proto/atomix/map/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/set,plugins=grpc:proto proto/atomix/set/*.proto
protoc -I=$proto_imports --go_out=Matomix/headers/headers.proto=github.com/atomix/atomix-go-node/proto/atomix/headers,import_path=atomix/value,plugins=grpc:proto proto/atomix/value/*.proto

proto_imports="./pkg:${GOPATH}/src/github.com/google/protobuf/src:${GOPATH}/src"

protoc -I=$proto_imports --go_out=import_path=atomix/list,plugins=grpc:pkg pkg/atomix/list/*.proto
protoc -I=$proto_imports --go_out=import_path=atomix/lock,plugins=grpc:pkg pkg/atomix/lock/*.proto
protoc -I=$proto_imports --go_out=import_path=atomix/map,plugins=grpc:pkg pkg/atomix/map/*.proto
protoc -I=$proto_imports --go_out=import_path=atomix/service,plugins=grpc:pkg pkg/atomix/service/*.proto