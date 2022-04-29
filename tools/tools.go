//go:build tools
// +build tools

package tools

import (
    _ "github.com/golang/mock/mockgen@latest"
    _ "github.com/sanposhiho/gomockhandler@latest"
    _ "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    _ "google.golang.org/protobuf/cmd/protoc-gen-go@latest"
)
