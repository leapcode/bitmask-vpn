# headless mode

As a wise person once said, "you don't want to struggle with Qt every day".

## backend

There's a barebones binary that launches the same backend that the qt5 client uses.

You will need a `providers.json` file containing the parameters for you own deployment. This is usually generated during the vendoring step, but you can manually edit the one for riseup:

```
go build ./cmd/bitmaskd
```


You might need to install the helpers (bitmask-root, polkit policies etc...). Do it manually, or use the embedded files (It will ask for sudo).

```
./bitmaskd -i
```


With the polkit files in place, you can now run bitmask backend in the foreground:

```
./bitmaskd -d gui/providers/providers.json 
```

TODO: make it a proper daemon, logging etc.

If you find problems while running (like polkit asking for password every time), you probably need to debug your polkit installation. Every system has its quirks, and bitmask has mostly been tested in debian-based desktops. For arch, you might need to add your user to group wheel.

## firewall

While testing, you are likely to get the iptables firewall leaving you with blocked outgoing connections. You can control `bitmask-root` manually:

```
sudo /usr/sbin/bitmask-root help
sudo /usr/sbin/bitmask-root firewall stop
```

## cli

There's no cli at the moment, but you can use the web api. To authenticate, you need to pass a token that is writen to a temporary file when the backend is initialized:

```
curl -H "X-Auth-Token:`cat /tmp/bitmask-token`" http://localhost:8000/vpn/status
curl -H "X-Auth-Token:`cat /tmp/bitmask-token`" http://localhost:8000/vpn/start
curl -H "X-Auth-Token:`cat /tmp/bitmask-token`" http://localhost:8000/vpn/stop
```
