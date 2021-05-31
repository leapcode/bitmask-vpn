Censorship Circumvention
================================================================================

This document contains some advice for using BitmaskVPN for censorship
circumvention.

Bootstrapping the connection
-----------------------------

There are two different steps where circumvention can be used: boostrapping the
connection (getting a certificate and the configuration files) and using an
obfuscated transport protocol. 

For the initial bootstrap, there are a couple of techniques that will be
attempted. If this fails, please open an issue with the relevant log
information.

Obfuscated bridges
-----------------------------

At the moment RiseupVPN offers obfs4 transport "bridges" (you can try them with
the `--obfs4` command line argument, or by checking the "use obfs4 bridges" checkbox on the preferences panel.

If you know you need bridges but the current ones do not work for you, please
get in contact. We're interested in learning what are the specific censorship
measures being deployed in your concrete location, and we could work together
to enable new bridges.

Getting certificates off-band
-----------------------------

As a last resort, you can place a valid certificate in the config folder (name
it after the provider domain). You might have downloaded this cert with Tor,
using a socks proxy etc...

  ~/.config/leap/riseup.net.pem

When the certificate expires you will need to download a new one.
