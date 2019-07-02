Branding for BitmaskVPN
================================================================================

This folder contains everything that is needed to generate a customized built of
BitmaskVPN for your provider.


Configure
--------------------------------------------------------------------------------

- Edit the file at 'branding/config/vendor.conf'. Add all the needed variables.
- Copy your provider CA certificate to 'branding/config/<provider>-ca.crt'
- Make sure that the folder 'branding/assets/<provider>' exists. Copy there all the needed assets.

Build
--------------------------------------------------------------------------------

Run:

PROVIDER=example.org make generate
make build
