#!/bin/bash

#sudo apt install upx

export GOPROXY=https://goproxy.io,direct
GOOS=linux go build -ldflags="-s -w" -o bin/mailsender src/main.go
GOOS=windows go build -ldflags="-s -w" -o bin/mailsender.exe src/main.go
upx bin/mailsender
