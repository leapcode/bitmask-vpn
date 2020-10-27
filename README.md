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
