Build
-----

Install dependencies:

```
  sudo make depends
```

Build the application:

```
  git clone 0xacab.org/leap/bitmask-vpn && cd bitmask-vpn
  make build
```

You need at least go 1.11. If you have something older and are using ubuntu, you can do:

```
  make install_go
```

For other situations, have a look at https://github.com/golang/go/wiki/Ubuntu or https://golang.org/dl/


OSX
---

You can install dependencies with homebrew:

```
  git clone 0xacab.org/leap/bitmask-vpn && cd bitmask-vpn
  make depends
  make build
```

Test
----

```
  sudo apt install qml-module-qttest
  make test
  make test_ui
```


Translations
------------

We use [transifex](https://www.transifex.com/otf/bitmask/RiseupVPN/) to coordinate translations. Any help is welcome!


Bugs?
-----

When you report an issue include the following information:

* what you expected to see
* what you got
* the version of the program. You can check the version on the about page.
* the logs of the program. The location of the logs depends on the OS:
  * linux: `/home/<your user>/.config/leap/systray.log`
  * OSX: `/Users/<your user>/Library/Preferences/leap/systray.log`, `/Applications/RiseupVPN.app/Contents/helper/helper.log` & `/Applications/RiseupVPN.app/Contents/helper/openvpn.log`
  * windows: `C:\Users\<your user>\AppData\Local\leap\systray.log`, `C:\Program Files\RiseupVPN\helper.log` & `C:\Program Files\RiseupVPN\openvp.log`
