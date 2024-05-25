## Supported operating systems

**BitmaskVPN** needs the following minimum versions of supported operating systems:

### On Windows

**BitmaskVPN** has been tested to work on windows 10 and 11 it might not work on earlier version of windows.

### On MacOS

- **BitmaskVPN** has been tested to work on last three releases of MacOS (Monteray, Ventura and Sonoma)
- **BitmaskVPN** currently needs rossetta to be enable to work on Apple hardware (M1, M2)

### On Linux

- **BitmaskVPN** has been tested to work on the latest version of Debian, Ubuntu, Fedora and Arch Linux
- Packages are only available for Ubuntu, Debian and Arch Linux

## Install

# arch

[There's a package in AUR](https://aur.archlinux.org/packages/riseup-vpn-git) that tracks main branch, so expect some instabilities (early birds catch the bugs they say, and we're thankful for that)

```
yaourt -Sy riseup-vpn-git
```

# deb

We haven't updated deb.leap.se repo yet 😞 (see #466), but if you *really* desire a debian
package you can build your own for the time being:

```
debuild -us -uc
sudo dpkg -i ../riseup-vpn*.deb
```

# ubuntu

If you're using ubuntu, you can use [leapcodes ppa](https://launchpad.net/~leapcodes/+archive/ubuntu/riseup-vpn).

```
sudo add-apt-repository ppa:leapcodes/riseup-vpn
sudo apt update
sudo apt install riseup-vpn
```

## Build

Clone this repo, install dependencies and build the application. Dependencies
assume debian packages, or homebrew for osx. For Windows OS see corresponding section below. For other systems try
manually, or send us a patch.

```
  git clone git@0xacab.org:leap/bitmask-vpn.git && cd bitmask-vpn
  sudo make depends  # do not use sudo in osx 
  make build
```

You need at least go 1.20.

## Test

You can run some tests too.

```
  sudo apt install qml-module-qttest
  make test
  make test_ui
```

## Windows
As for now app can be build on Win OS using `Cygwin` terminal.

#### Precondition
You need to have installed and added to your user PATH (mentioned version tested in Win10):
1) Go (>= go1.20)
2) QT (>= Qt6.6)
3) QtIFW (>= QtIFW-4.0.0)
4) Cygwin64 (>= 2.905 64 bit)
5) Using Cygwin `Package Select` window install `python3` and `make` packages. 

**Note:** for \#5 you don't need to add packages to PATH they will available in `cygwin` after installation.

#### Get Source
```
    git clone git@0xacab.org:leap/bitmask-vpn.git && cd bitmask-vpn
```

#### Build
Build script uses a symbolic link in one of the stages. Unfortunately Cygwin can't create native symlink from local non   
admin user due to windows security restriction. To avoid this issue we need to call next target from cygwin terminal as   
Administrator. This need to be done only once. 
```bash
    make relink_vendor
```

After `relink_vendor` use this to build the app:
```bash
    make build
```
After successful build application will be available at: `build/qt/release/riseup-vpn.exe`

#### Test

To run tests:

```bash
    make test
    make test_ui
```

## Logging

Log files:
Linux: `~/.config/leap/systray.log`
Windows: `%LocalAppData%\leap\systray.log  `
Mac: `~/Library/Preferences/leap/systray.log`

Log levels can be set via environment variable (`LOG_LEVEL=TRACE`, `LOG_LEVEL=DEBUG`, default log level is `INFO`). The cpp/qml part logs to stderr if env `DEBUG=1` is set.

Translations
------------

We use [transifex](https://www.transifex.com/otf/bitmask/bitmask-desktop/) to coordinate translations. Any help is welcome!


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
