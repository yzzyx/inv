FROM fedora

RUN dnf -y install mingw64-gtk3 go mingw32-binutils mingw32-nsiswrapper
RUN bash -c "sed -i -e 's/-Wl,-luuid/-luuid/g' /usr/x86_64-w64-mingw32/sys-root/mingw/lib/pkgconfig/gdk-3.0.pc"
RUN bash -c "PKG_CONFIG_PATH=/usr/x86_64-w64-mingw32/sys-root/mingw/lib/pkgconfig CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go get -u github.com/gotk3/gotk3/gtk"
RUN bash -c "PKG_CONFIG_PATH=/usr/x86_64-w64-mingw32/sys-root/mingw/lib/pkgconfig CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go install -v github.com/gotk3/gotk3/gtk"

COPY "compile.sh" /bin
ENTRYPOINT "compile.sh"
