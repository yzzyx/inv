#!/bin/bash
set -e

cd /build
export PKG_CONFIG_PATH=/usr/x86_64-w64-mingw32/sys-root/mingw/lib/pkgconfig
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export GOOS=windows
export GOARCH=amd64
go build -ldflags -H=windowsgui
