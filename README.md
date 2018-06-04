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
  $ git clone 0xacab.org/leap/bitmask-systray
  $ cd bitmask-systray
  $ go get .
  $ go build
```

Run it
-------------
bitmask-systray assumes that you already have bitmaskd running.

Run bitmask and the systray:
```
  $ bitmaskd
  $ ./bitmask-systray
```

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
