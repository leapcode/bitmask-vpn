BitmaskVPN Branding Procedure
================================================================================

This document contains the instructions to generate a custom build of the
Desktop BitmaskVPN app for a given provider.

Configure
--------------------------------------------------------------------------------

All the needed information to vendorize BitmaskVPN are contained in an external
folder, where you will place the connection details to your own provider and
any asset that you want to customize. To start a new vendoring project, you need
to initialize a new repo for your provider:

  export VENDOR_PATH=../leapvpn-myprovider-pkg
  make vendor_init

Follow the directions in the output of the above command. Basically you need to
configure your provider CA certificate, and some graphical assets:

  * Copy your provider CA certificate to the same folder: '<provider>-ca.crt'
  * Check the list of assets at 'assets/FILES.Readme'.

You can validate your configuration:

  VENDOR_PATH=../myprovider-vpn-pkg vendor_check

This will fetch your provider's CA against the one you have configured. If you
want to skip the online check, set the `SKIP_CACHECK` to "yes".

Checkout the source
--------------------------------------------------------------------------------

 git clone https://0xacab.org/leap/bitmask-vpn
 cd bitmask-vpn
 git pull --tags


Build & package
--------------------------------------------------------------------------------

NOTE: Some of the following scripts need network access, since they will check
whether the configuration published by your provider matches what is configured
before the build. If you want to skip this check, pass `SKIP_CACHECK=yes`

Run::

 VENDOR_PATH=../myprovider-vpn-pkg make vendor

Then you can build the binaries for some quick manual testing::

 make build

Now you can build the installer for your host platform::

 make build_installer

Previously we had a cross-compilation setup in place. Cross compilation will be added back in the future.

For debian and snap packages (FIXME -- WORK IN PROGRESS):

  make debian
  make snap
