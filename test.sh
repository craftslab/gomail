#!/bin/bash

export GOPROXY=https://goproxy.io,direct
go test -v src/*.go
