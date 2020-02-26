#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
go vet sender/*.go
go vet parser/*.go
gofmt -s -w sender/*.go
gofmt -s -w parser/*.go
