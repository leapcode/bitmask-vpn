Install it
----------

Install dependencies:
```
  # apt install libzmq3-dev libgtk-3-dev libappindicator3-dev golang
```

Build the systray:
```
  $ git clone 0xacab.org/leap/bitmask-systray
  $ cd bitmask-systray
  $ go get .
  $ go build
```

Run bitmask and the systray:
```
  $ bitmaskd
  $ ./bitmask-systray
```
