Install it
----------

Install dependencies:
```
  # apt install libzmq3-dev libgtk-3-dev libappindicator3-dev golang pkg-config
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

OSX
----------
Using homebrew:

```
  $ brew install golang zmq pkg-config
  $ git clone 0xacab.org/leap/bitmask-systray
  $ cd bitmask-systray
  $ go get .
  $ go build
```

Run it
-------------
bitmask-systray assumes that you already have bitmaskd running.
