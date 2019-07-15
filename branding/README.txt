Branding for BitmaskVPN
================================================================================

This folder contains everything that is needed to generate a customized built of
BitmaskVPN for your provider.


Configure
--------------------------------------------------------------------------------

- Copy or edit the file at 'branding/config/vendor.conf'. Add all the needed variables.
- Copy your provider CA certificate to the same folder: 'branding/config/<provider>-ca.crt'
- Make sure that the folder 'branding/assets/<provider>' exists. Copy there all the needed assets.

Checkout
--------------------------------------------------------------------------------

git clone https://0xacab.org/leap/bitmask-vpn
cd bitmask-vpn
git pull --tags

Build
--------------------------------------------------------------------------------

make build


Package
--------------------------------------------------------------------------------

NOTE: Some of the following scripts need network access, since they will check
whether the configuration published by your provider matches what is configured
before the build.

Run:

PROVIDER=example make prepare_all

You can also specify a cusom config file:

PROVIDER=example PROVIDER_CONFIG=/path/to/vendor.conf make prepare_all
make build

After this, you will find the build scripts ready in the following folder:

cd build/example

make package_win
make package_osx
make package_snap
make package_deb

