Install it
----------

Install dependencies:

TODO: add qt5 deps here

```
  # make depends
```

Build the systray:
```
  $ git clone 0xacab.org/leap/bitmask-vpn && cd bitmask-vpn
  $ make build
```

You need at least go 1.11. If you have something older and are using ubuntu, you can do:

```
  make install_go
```

For other situations, have a look at https://github.com/golang/go/wiki/Ubuntu or https://golang.org/dl/


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

./build.sh


i18n
----

TODO: move this to developer docs

The translations are done in transifex. To help us contribute your translations there and/or review the existing
ones:
https://www.transifex.com/otf/bitmask/RiseupVPN/

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
  * linux: `/home/<your user>/.config/leap/systray.log`
  * OSX: `/Users/<your user>/Library/Preferences/leap/systray.log`, `/Applications/RiseupVPN.app/Contents/helper/helper.log` & `/Applications/RiseupVPN.app/Contents/helper/openvpn.log`
  * windows: `C:\Users\<your user>\AppData\Local\leap\systray.log`, `C:\Program Files\RiseupVPN\helper.log` & `C:\Program Files\RiseupVPN\openvp.log`
