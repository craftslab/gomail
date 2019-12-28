#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
go vet src/*.go
go vet tools/recipientparser/*.go
