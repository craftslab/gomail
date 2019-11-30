#!/bin/bash

export GOPROXY=https://goproxy.io,direct
go mod tidy
go vet src/*.go
