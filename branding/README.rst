BitmaskVPN Branding Procedure
================================================================================

This folder contains everything that is needed to generate a customized built of
the Desktop BitmaskVPN app for a given provider.


Configure
--------------------------------------------------------------------------------

To start a new vendoring project, initialize a new repo for your provider:

  export VENDOR_PATH=../leapvpn-myprovider-pkg
  make vendor_init

Follow the directions in the output of the above command. Basically you need to
configure your provider CA certificate, and some graphical assets:

  * Copy your provider CA certificate to the same folder: 'config/<provider>-ca.crt'
  * Check the list of assets at 'assets/FILES.Readme'.

You can validate your configuration:

  export VENDOR_PATH=../leapvpn-myprovider-pkg
  make vendor_check

Checkout
--------------------------------------------------------------------------------

 git clone https://0xacab.org/leap/bitmask-vpn
 cd bitmask-vpn
 git pull --tags


Package
--------------------------------------------------------------------------------

NOTE: Some of the following scripts need network access, since they will check
whether the configuration published by your provider matches what is configured
before the build. If you want to skip this check, pass `SKIP_CACHECK=yes`

Run::

 export VENDOR_PATH=../leapvpn-myprovider-pkg
 make vendor
 make prepare

Then you can build the binary::

 make build

* FIXME: the following does not work yet ---------------------
  REFACTORING in progress ------------------------------------

Then you can build all the packages::

 make packages

Alternatively, you can build only for an specific os::

 make package_win
 make package_osx
 make package_snap
 make package_deb
