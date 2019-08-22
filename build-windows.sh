#!/bin/bash

if [ ! -d "gtk-win64-runtime" ]; then
    echo "Download runtime from https://github.com/tschoonj/GTK-for-Windows-Runtime-Environment-Installer and install to folder 'gtk-win64-runtime'"
    exit 1
fi

#export PKG_CONFIG_PATH=/usr/x86_64-w64-mingw32/lib/pkgconfig
export PKG_CONFIG_ALLOW_CROSS=1
export PKG_CONFIG_PATH=/mingw64/lib/pkgconfig
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export GOOS=windows
export GOARCH=amd64
go install github.com/gotk3/gotk3/gtk

go build -ldflags -H=windowsgui
mkdir -p windows
mkdir -p windows/share/glib-2.0/schemas
mkdir -p windows/share/icons

cp gtk-win64-runtime/*.dll windows
cp -r gtk-win64-runtime/share windows
cp -r gtk-win64-runtime/etc windows
cp -r gtk-win64-runtime/lib windows
cp inv.glade windows
cp inv.exe windows
