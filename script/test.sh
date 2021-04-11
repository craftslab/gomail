#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

if [ "$1" = "all" ]; then
  GIN_MODE=release go test -cover -v -parallel 2 ./...
else
  list="$(go list ./... | grep -v foo)"
  old=$IFS IFS=$'\n'
  for item in $list; do
    GIN_MODE=release go test -cover -v -parallel 2 "$item"
  done
  IFS=$old
fi
