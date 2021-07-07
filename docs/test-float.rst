Testing float deployments
=========================
You can quickly brand the client for test float instances.::

  export PROVIDER=floatdemo
  make vendor && make build
  build/qt/release/floatdemo-vpn

If your test instance is not there, just add it to `providers/vendor.conf`.
