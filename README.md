# Bitmask - Desktop client

## Supported operating systems

**Bitmask** needs the following minimum versions of supported operating systems:

### On Windows

**Bitmask** has been tested to work on windows 10 and 11 it might not work on earlier version of windows.

### On MacOS

- **Bitmask** has been tested to work on last three releases of MacOS (Monteray, Ventura and Sonoma)
- **Bitmask** currently needs rossetta to be enable to work on Apple hardware (M1, M2)

### On Linux

- **Bitmask** has been tested to work on the latest version of Debian, Ubuntu, Fedora and Arch Linux
- Packages are only available for Ubuntu, Debian and Arch Linux

## Install

## Build

Clone this repo, install dependencies and build the application. Dependencies assume debian packages, or homebrew for osx. For Windows OS see corresponding section below. For other systems try manually, or send us a patch. bitmask-vpn can be branded for a specific provider by specifying the env variable PROVIDER during the build process; we currently support three providers: riseup, calyx, and bitmask. To create a client branded for 'riseup', run:

```
git clone git@0xacab.org:leap/bitmask-vpn.git && cd bitmask-vpn
sudo make depends  # do not use sudo in osx 
PROVIDER=riseup make vendor
make build
sudo build/qt/release/riseup-vpn --install-helpers # on Linux and Mac
LOG_LEVEL=TRACE build/qt/release/riseup-vpn
```

With `--install-helpers` the `bitmask-root` helper gets copied to `/usr/sbin`.

# Ubuntu

If you're using Ubuntu, you can use [leapcodes ppa](https://launchpad.net/~leapcodes/+archive/ubuntu/riseup-vpn).

```
sudo add-apt-repository ppa:leapcodes/riseup-vpn
sudo apt update
sudo apt install riseup-vpn
```

# Debian

The package is available as "riseup-vpn" in Debian Bookworm, albeit at an older version. To get the same, you could run:

```
sudo apt install riseup-vpn
```

The latest version is available for Debian Bookworm via backports. See the [offcial page](https://backports.debian.org/Instructions/) for instructions on how to set it up. If you are using Debian Testing/Unstable, riseup-vpn's latest version is available there as well.

If you're using an older version of Debian, then we do not have a package for the same. However, if you really desire a debian package you can build your own for the time being:

```
debuild -us -uc
sudo dpkg -i ../riseup-vpn*.deb
```

You can also run 
```
PROVIDER=riseup make vendor
PROVIDER=riseup QMAKE=qmake6 make package_deb
```
Then install the built package with `apt install -f ./deploy/*.deb`.


# Arch Linux

There are two AUR packages for Arch Linux. There is [riseup-vpn-git](https://aur.archlinux.org/packages/riseup-vpn-git) that tracks main branch, so expect some instabilities (early birds catch the bugs they say, and we're thankful for that). There is also [riseup-vpn](https://aur.archlinux.org/packages/riseup-vpn) with the latest stable release.

```
yay riseup-vpn
```

## Snap

There is also a package in the [Snap store](https://snapcraft.io/riseup-vpn).

```
sudo snap install riseup-vpn
```

## Build

Clone this repo, install dependencies and build the application. Dependencies assume debian packages, or homebrew for osx. For Windows OS see corresponding section below. For other systems try manually, or send us a patch. bitmask-vpn can be branded for a specific provider by specifying the env variable PROVIDER during the build process; we currently support three providers: riseup, calyx, and bitmask. To create a client branded for 'riseup', run:

```
git clone git@0xacab.org:leap/bitmask-vpn.git && cd bitmask-vpn
sudo make depends  # do not use sudo in osx 
PROVIDER=riseup make vendor
make build
```

To build you need at least go 1.22.

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

Log levels can be set via environment variable (`LOG_LEVEL=TRACE`, `LOG_LEVEL=DEBUG`, default log level is `INFO`). The cpp/qml part logs to stderr if env `DEBUG=1` is set. If `OPENVPN_LOG_TO_FILE=1` is set, the OpenVPN process writes its logs to [os.TempDir()](https://pkg.go.dev/os#TempDir)/leap-vpn.log. The verbosity of OpenVPN can be specified with env `OPENVPN_VERBOSITY` (sets `--verb`).

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
