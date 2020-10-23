some attempts at mxe cross-compilation
======================================

We really would like to have cross-compilation working.

I tried mxe, it looks the most promising way to, at least, get the binaries
working. (Cross-compiling a static version of QtIFW might prove more difficult,
though).

these two links were useful for me in my attempts:

https://gist.github.com/amitsaha/ec8fbbc01e22ef9cc020570f415fa2fb
https://stackoverflow.com/questions/14170590/building-qt-5-on-linux-for-windows

I tried the mxe project stretch packages

* debs seem to be broken :(
* add this repo::

  deb http://pkg.mxe.cc/repos/apt stretch main

* install this package::

  mxe-x86_64-w64-mingw32.static-qt5


- I think I tried with cmake. should try again now that I went the qmake route.
- Compiling things with the instructions above got me further. However, I only compiled a very simple qt app - did not try with all the QML/foo libraries. It should not be much harder...
- I had to patch some files in mxe to workaround a couple of issues (basically editing include paths). TODO -- dig those patches and include them here.
