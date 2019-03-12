Install it
----------

Install dependencies:
```
  # apt install libgtk-3-dev libappindicator3-dev golang pkg-config
```

Build the systray:
```
  $ git clone 0xacab.org/leap/bitmask-vpn
  $ cd bitmask-systray
  $ go get .
  $ go build
```

To be able to build the assets you'll need:
```
  $ go get -u golang.org/x/text/cmd/gotext github.com/cratonica/2goarray
```

OSX
----------
Using homebrew:

```
  $ brew install golang zmq pkg-config
  $ brew install --default-names gnu-sed
  $ git clone 0xacab.org/leap/bitmask-vpn
  $ cd bitmask-systray
  $ go get .
  $ go build

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


Run it
-------------
The default build is a standalone systray. It still requires a helper and openvpn installed to work. For linux the helper is
[bitmask-root](https://0xacab.org/leap/bitmask-dev/blob/master/src/leap/bitmask/vpn/helpers/linux/bitmask-root)
for windows and OSX there is [a helper written in go](https://0xacab.org/leap/riseup_vpn/tree/master/helper).

To build and run it:
```
  $ go build
  $ ./bitmask-systray
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
  $ ./bitmask-systray
```


i18n
----

When a string has being modified you need to regenerate the locales:
```
  $ make generate_locales
```

To fetch the translations from transifex and rebuild the catalog.go (API\_TOKEN is the transifex API token):
```
  $ API_TOKEN='xxxxxxxxxxx' make locales
```
There is some bug on gotext and the catalog.go generated doesn't have a package, you will need to edit
cmd/bitmask-vpn/catalog.go and to have a `package main` at the beginning of the file.

If you want to add a new language create the folder `locales/$lang` before running `make locales`.


Report an issue
-------------------

When you report an issue include the following information:
* what you expected to see
* what you got
* the version of the program. You can check the version on the about page.
* the logs of the program. The location of the logs depends on the OS:
  * linux: `/home/<your user>/.config/leap/bitmaskd.log` & `/home/<your user>/.config/leap/systray.log`
  * OSX: `/Users/<your user>/Library/Preferences/leap/systray.log`, `/Applications/RiseupVPN.app/Contents/helper/helper.log` & `/Applications/RiseupVPN.app/Contents/helper/openvpn.log`
  * windows: `C:\Users\<your user>\AppData\Local\leap\systray.log`, `C:\Program Files\RiseupVPN\helper.log` & `C:\Program Files\RiseupVPN\openvp.log`
