git integration
===============

All seems to be more smooth with the "new" (ahem) github integration (once things *are* working).

Some tips:

- We've got different repos. `riseup-vpn-snap` is the *snap* repo. Import code from upstream, just merge it with `-X theirs`
- If the snap doesn't change, just use `make bump_snap` for upgrading the version from git (TODO we could write this also into the hardcoded version).
- Otherwise, just do `make vendor` and import the snapcraft.yaml generated from the template.


local builds
------------
multipass is the recommended way, but canonical does use lxd so at times some paths etc change.
For your own sanity:

- get a zfs pool on a fast device, and get yourself acquainted with lxd to use that pool.
- don't get too frustrated with networking + lxd. restarting any iptables in your host (if using bridges) usually helps.
- you can use `make local_snap` to use your local lxd infra. it launches with
  `--debug`, so you'll be dropped into a local shell to see what the fuck the
  manual build of Qt is complaining about.


existential helpline
--------------------
* don't despair. we've all been there.
* snapcraft forum is useful.
* all tech is crap: don't think that you'll be happy reimplementing the whole
  app in electron or whatnot. just don't. enjoy life while you can.
* https://forum.snapcraft.io/t/the-sorry-state-of-snapping-qt5-apps/22809
* https://github.com/mozilla-mobile/mozilla-vpn-client/blob/main/scripts/qt5_compile.sh

if you have some time
---------------------
* look into a `clang` build. qt builds fine, but last time I tried there was
  some incompatible version (?) that didn't let the qmake build finish.

