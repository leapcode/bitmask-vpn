windows build
=============

The build currently expects MINGW64 environment, on a native windows host.
A cross-compiling procedure (at least for the application binaries) should be possible in the near future, using mxe. (There's already some support for it in `gui/build.sh`).

You should instal: make, wget, as well as a recent Qt5 version (for instance, with chocolatey: choco install make && choco install wget).

(In order to avoid makefiles, you are welcome to submit a port of the build scripts using powershell or cscript - see the build.wsf script in openvpn-build for inspiration).

For the installer, install QtIFW for windows (tested with version 3.2.2).

Assuming you have the vendor path in place and correctly configured, all you need to do is `make build_installer`::

  export PATH="/c/Qt/Qt5/bin/":"/c/Qt/QtIFW-3.2.2/bin":$PATH
  VENDOR_PATH=providers/
  make build_installer
