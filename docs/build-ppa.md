# PPA How to

LEAP team maintains a [ppa repository](https://launchpad.net/~leapcodes) for the clients, pacakges are built for latest two LTS releases of ubuntu

## Pre-requisites

Ensure that all the build dependencies are already installed, you can use `make depends` on most ubuntu and debian version to have the machine
ready to build `bitmask-vpn` debian packages

If `make depends` do not work, it is useful to have the `devscripts` and `equivs` packages installed, these are needed later for building
the source package and installing build dependencies.

PPA expects a signed source package, we have to build this package and then upload to PPA the changes file using the [`dput`](https://manpages.ubuntu.com/manpages/xenial/man1/dput.1.html) tool.

Please refer to official [PPA documentation](https://help.launchpad.net/Packaging/PPA) for how to create an account and add SSH and GPG keys to be able to upload.

## Build signed source package

### Prepare the debian package from templates

```
$ export PROVIDER=riseup # can be riseup, bitmask or calyx
$ make vendor
$ BUILD_RELEASE=yes make prepare_deb
```

> **NOTE**: The above commands will generate a debian directory in `build/riseup/debian` the control file created there can be used to build a dependencies package

* If build depends are not yet installed, build a dependencies package with all the build and runtime dependencies of `bitmask-vpn`:

```
$ cd build/riseup/debian
$ mk-build-deps control
$ apt-get install -f ./riseup-vpn-build-deps_0.24.8_all.deb
```

* Add changes to changelog by copying the entries from the `CHANGELOG` file at the root of the repo

```
# example changelog file for 0.24.8 might look like
$ cd build/riseup/build/riseup-vpn_0.24.8/
$ cat debian/changelog
riseup-vpn (0.24.8~noble) noble; urgency=medium
  * Reduces the size of splash screen image
  * Disable obfs4 and kcp checkbox in preferences for riseup
  * Removes duplicate languages in the language picker in preferences
  * Language picker in preferences shows languages sorted alphabetically
  * 0.24.8 ubuntu noble release

 -- LEAP Encryption Access Project <debian@leap.se>  Thu, 05 Sep 2024 03:06:54 +0800

riseup-vpn (0.24.8-6-g92db03c4) unstable; urgency=medium

  * Initial package.

 -- LEAP Encryption Access Project  <debian@leap.se>  Mon, 29 Jul 2019 10:00:00 +0100

```

* Bump native dot-version, change release

```
$ cd build/riseup/build/riseup-vpn_0.24.8
# to add a new entry for version 0.24.8 to the changelog file and update the release
$ dch -b -v 0.24.8~noble -D "noble" -m "riseup-vpn release 0.24.8"
```

> **NOTE:** The source tarball's name as set by the `make preapre_deb` step will not match the version we set in the changelog file, since
for PPAs we need to append the distribution name to the version, e.g to build `0.24.8` for `noble` the version is `0.24.8~noble`
> More details about versioning ppa can be found in the PPA docs [versioning section](https://help.launchpad.net/Packaging/PPA/BuildingASourcePackage#versioning)

* We need to rename the source tarball to match the version we set in the `changelog` file:

```
$ cd build/riseup/build
$ mv riseup-vpn_0.24.8.orig.tar.gz riseup-vpn_0.24.8~noble.orig.tar.gz
```

### Build signed source package

```
$ cd build/riseup/build/riseup-vpn_0.24.8
$ debuild -S -k=<key_id_for_signing>
```

### Upload changes file

```
$ cd build/riseup/build
$ dput ppa:leapcodes/ppa riseup-vpn_0.24.8~noble_source.changes
```
