// +build tools

package main

import (
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/jteeuwen/go-bindata/go-bindata"
)
