Branding for BitmaskVPN
================================================================================

This folder contains everything that is needed to generate a customized built of
the Desktop BitmaskVPN app for a given provider.


Configure
--------------------------------------------------------------------------------

* Copy or edit the file at 'branding/config/vendor.conf'. Add all the needed variables.
* Copy your provider CA certificate to the same folder: 'branding/config/<provider>-ca.crt'
* Make sure that the folder 'branding/assets/<provider>' exists. Copy there all the needed assets.

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

 PROVIDER=example make prepare

You can also specify a custom config file::

 PROVIDER=example PROVIDER_CONFIG=/path/to/vendor.conf make prepare

Then you need to build the package::

 make build

Then you can build all the packages::

 make packages

Alternatively, you can build only for an specific os::

 make package_win
 make package_osx
 make package_snap
 make package_deb
