#!/bin/bash

go env -w GOPROXY=https://goproxy.cn,direct

CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/mailsender src/sender.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/recipientparser tools/recipientparser/parser.go

CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o bin/mailsender.exe src/sender.go
CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -ldflags="-s -w" -o bin/recipientparser tools/recipientparser/parser.go

apt install upx

upx bin/mailsender
upx bin/mailsender.exe

upx bin/recipientparser
upx bin/recipientparser.exe
