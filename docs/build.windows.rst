windows build
=============

The build currently expects MINGW64 environment, on a native windows host.
A cross-compiling procedure (at least for the application binaries) should be possible in the near future, using mxe. (There's already some support for it in `gui/build.sh`).

You should instal: make, wget, as well as a recent Qt5 version (for instance, with chocolatey: choco install make && choco install wget).

(In order to avoid makefiles, you are welcome to submit a port of the build scripts using powershell or cscript - see the build.wsf script in openvpn-build for inspiration).

For the installer, install QtIFW for windows (tested with version 3.2.2).

Assuming you have the vendor path in place and correctly configured, all you need to do is `make build_installer`::

  export PATH="/c/Qt/Qt5/bin/":"/c/Qt/QtIFW-3.2.2/bin":$PATH
  export VENDOR_PATH=providers
  make vendor && make installer


checking signatures
-------------------
we should be signing all binaries on a release build.

to check the binaries have proper signatures, you can use the sigcheck
utilities, part of the sysinternals suite:

https://docs.microsoft.com/en-us/sysinternals/downloads/sysinternals-suite

unzip and place sigcheck.exe somewhere in your path.

make sure to pass -accepteula parameter on some manual run so that it does not
ask again.

adding metadata to binaries
---------------------------
TODO: add metadata properly
https://github.com/electron/rcedit/releases/tag/v1.1.1
https://stackoverflow.com/questions/284258/how-do-i-set-the-version-information-for-an-existing-exe-dll

the steps to do release signatures are::

  make build
  make dosign
  make installer
  make sign_installer

or alltogether as::

  make package_win_release


unreviewed notes
----------------
see comment about patching dlls and windeployqt not being needed anymore https://stackoverflow.com/a/61910592
