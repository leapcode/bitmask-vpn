Install it
----------

Install dependencies:
```
  # apt install libgtk-3-dev libappindicator3-dev golang pkg-config
```

Build the systray:
```
  $ git clone 0xacab.org/leap/bitmask-systray
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
  $ git clone 0xacab.org/leap/bitmask-systray
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
bitmask-systray assumes that you already have bitmaskd running.

Run bitmask and the systray:
```
  $ bitmaskd
  $ ./bitmask-systray
```

Standalone
-------------

Is also posible to compile the systray to be standalone (don't depend on bitmask):
```
  $ go build -tags standalone
```
It still requires a helper and openvpn installed to work. For linux the helper is
[bitmask-root](https://0xacab.org/leap/bitmask-dev/blob/master/src/leap/bitmask/vpn/helpers/linux/bitmask-root)
for windows and OSX there is [a helper written in go](https://0xacab.org/leap/riseup_vpn/tree/master/helper).


i18n
----

Generate `locales/*` files:
```
  $ make generate_locales LANGS="sjn tlh"
```

Edit the `locales/*/out.gotext.json` translations into `locales/*/messages.gotext.json`.

To rebuild the locales:
```
  $ make locales
```
