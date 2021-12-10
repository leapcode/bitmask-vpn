# I have time, how can I help?

## Packaging

* Look into `AppImage` + https://github.com/probonopd/linuxdeployqt.
  We've not considered that option too much in the past, but it might give us
  a decent, self-contained alternative to snap etc.
 
## Linux

* Revamp vpn helper architecture: there're problems, of course, but we can try
  to isolate the client gui from the vpn helper itself (and "ship it" as
  a binary under the single bundle, as I do now with bitmak-root). An idea
  that's been floating around for a long time is to recycle the helper
  interface, and have a long-lived privileged helper that does the vpn
  connection using openvpn3 libr.

  elijah was initially supporting a short-lived helper (what we have right now
  with bitmask-root), but perhaps the integration can be done right with pkexec
  or otherwise (separate users in linux etc). This has the additional advantage
  of allowing us to do a very early startup, and not to depend so much on
  pkexec + ubuntu's quirks (portability!).
