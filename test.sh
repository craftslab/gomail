#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct
go test -v src/*.go
