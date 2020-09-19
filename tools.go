// +build tools

package main

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
	_ "github.com/tmthrgd/go-bindata/go-bindata"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
