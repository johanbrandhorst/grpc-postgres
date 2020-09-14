// +build tools

package main

import (
	_ "github.com/bufbuild/buf/cmd/buf"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
	_ "github.com/tmthrgd/go-bindata/go-bindata"
)
