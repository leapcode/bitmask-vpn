Build
-----

Clone this repo, install dependencies and build the application. Dependencies
assume debian packages, or homebrew for osx. For other systems try
manually, or send us a patch.

```
  git clone 0xacab.org/leap/bitmask-vpn && cd bitmask-vpn
  sudo make depends
  make build
```

You need at least go 1.11. If you have something older and are using ubuntu, you can do:

```
  make install_go
```

For other situations, have a look at https://github.com/golang/go/wiki/Ubuntu or https://golang.org/dl/

Test

OSX
----------
Using homebrew:

```
  $ git clone 0xacab.org/leap/bitmask-vpn && cd bitmask-vpn
  $ make depends
  $ make build

```

Linux
----------
Building the systray in linux will produce some `-Wdeprecated-declarations` warnings, like that:
```
cgo-gcc-prolog: In function ‘_cgo_3f9f61f961c9_Cfunc_gtk_font_button_get_font_name’:
cgo-gcc-prolog:5455:2: warning: ‘gtk_font_button_get_font_name’ is deprecated [-Wdeprecated-declarations]
In file included from /usr/include/gtk-3.0/gtk/gtk.h:106:0,
                 from ../../../go/src/github.com/gotk3/gotk3/gtk/gtk.go:48:
/usr/include/gtk-3.0/gtk/gtkfontbutton.h:96:23: note: declared here
 const gchar *         gtk_font_button_get_font_name  (GtkFontButton *font_button);
                       ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~
```
They are expected and don't produce any problem on the systray.

Windows
---------
<<<<<<< HEAD
Download cygwinn // https://cygwin.com/setup-x86_64.exe
`````
Install with the necessary packages (my case 64bit):
=======
Download cygwin // https://cygwin.com/setup-x86_64.exe
```
Install with the necessary packages:

>>>>>>> 9901a4e... Readme update
mingw64-x86_64-gcc-core
mingw64-x86_64-gcc-g++ 
and
x86_64-w64-mingw32-c++
x86_64-w64-mingw32-gcc
make

<<<<<<< HEAD

````
Add to Windows Path "C:\cygwin64\bin"

=======
Add to windowspath "C:\cygwin64\bin"
```
Build it
```
make build 

Build flags
ARCH : 386 or amd64 (default: amd64)
CCPAath and CXXPath are either paths of compiler or filenames in %PATH% (defaults: x86_64-w64-mingw32-gcc and x86_64-w64-mingw32-c++)

Examples:
make build ARCH=386
make build ARCH=386 CCPath=i686-w64-mingw32-gcc CXXPath=i686-w64-mingw32-c++

All options can be omitted! 

```
>>>>>>> 9901a4e... Readme update

Run it
-------------
The default build is a standalone systray. It still requires a helper and openvpn installed to work. For linux the helper is
[bitmask-root](https://0xacab.org/leap/bitmask-dev/blob/master/src/leap/bitmask/vpn/helpers/linux/bitmask-root)
for windows and OSX there is [a helper written in go](https://0xacab.org/leap/bitmask-vpn/tree/master/pkg/helper/).

Run it:
```
  $ build/bin/bitmask-vpn

```


Bitmaskd
-------------
Is also posible to compile the systray to use bitmask as backend:
```
  $ go build -tags bitmaskd
```

In that case bitmask-systray assumes that you already have bitmaskd running. Run bitmask and the systray:
```
  $ bitmaskd
  $ build/bin/bitmask-vpn
```


i18n
----

You can run some tests too.

```
  sudo apt install qml-module-qttest
  make test
  make test_ui
```


Translations
------------

We use [transifex](https://www.transifex.com/otf/bitmask/RiseupVPN/) to coordinate translations. Any help is welcome!


Bugs? Crashes? UI feedback? Any other suggestions or complains?
---------------------------------------------------------------

When you are willing to [report an issue](https://0xacab.org/leap/bitmask-vpn/-/issues) please
use the search tool first. if you cannot find your issue, please make sure to
include the following information:

* the platform you're using and the installation method.
* the version of the program. You can check the version on the "about" menu.
* what you expected to see.
* what you got instead.
* the logs of the program. The location of the logs depends on the OS:
  * gnu/linux: `/home/<your user>/.config/leap/systray.log`
  * OSX: `/Users/<your user>/Library/Preferences/leap/systray.log`, `/Applications/RiseupVPN.app/Contents/helper/helper.log` & `/Applications/RiseupVPN.app/Contents/helper/openvpn.log`
  * windows: `C:\Users\<your user>\AppData\Local\leap\systray.log`, `C:\Program Files\RiseupVPN\helper.log` & `C:\Program Files\RiseupVPN\openvp.log`
