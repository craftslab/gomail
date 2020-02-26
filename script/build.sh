#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/mailsender sender/sender.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/recipientparser parser/parser.go

CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o bin/mailsender.exe sender/sender.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o bin/recipientparser.exe parser/parser.go

apt install upx

upx bin/mailsender
upx bin/mailsender.exe

upx bin/recipientparser
upx bin/recipientparser.exe
