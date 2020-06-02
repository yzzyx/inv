#!/bin/bash

if [ ! -d "gtk-win64-runtime" ]; then
    echo "Download runtime from https://github.com/tschoonj/GTK-for-Windows-Runtime-Environment-Installer and install to folder 'gtk-win64-runtime'"
    exit 1
fi

(cd docker/w64build && docker build . -t w64-go-cache)
docker run -v $(pwd):/build -v w64-go-cache:/root/go w64-gtk-build

mkdir -p windows
mkdir -p windows/share/glib-2.0/schemas
mkdir -p windows/share/icons

cp gtk-win64-runtime/*.dll windows
cp -r gtk-win64-runtime/share windows
cp -r gtk-win64-runtime/etc windows
cp -r gtk-win64-runtime/lib windows
cp inv.glade windows
cp inv.exe windows
