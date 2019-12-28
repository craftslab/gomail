#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
go test -cover -v src/*.go
go test -cover -v tools/recipientparser/*.go
