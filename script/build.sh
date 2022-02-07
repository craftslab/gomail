#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

ldflags="-s -w"
list="parser,sender"
old=$IFS IFS=$','

for item in $list; do
  CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "$ldflags" -o bin/$item $item/$item.go
  CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags "$ldflags" -o bin/$item.exe $item/$item.go
  upx bin/$item
  upx bin/$item.exe
done

IFS=$old
