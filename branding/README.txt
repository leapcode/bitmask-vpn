Branding for BitmaskVPN
================================================================================

This folder contains everything that is needed to generate a customized built of
BitmaskVPN for your provider.


Configure
--------------------------------------------------------------------------------

- Copy or edit the file at 'branding/config/vendor.conf'. Add all the needed variables.
- Copy your provider CA certificate to the same folder: 'branding/config/<provider>-ca.crt'
- Make sure that the folder 'branding/assets/<provider>' exists. Copy there all the needed assets.

Build
--------------------------------------------------------------------------------

Some of the following scripts need network access, since they will check
whether the configuration published by your provider matches what is configured
before the build.

Run:

PROVIDER=example.org make prepare
make build

You can also specify a cusom config file:

PROVIDER=example.org PROVIDER_CONFIG make prepare
make build


