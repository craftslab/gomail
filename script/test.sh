#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
go test -cover -v sender/*.go
go test -cover -v parser/*.go
