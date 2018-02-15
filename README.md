Install it
----------

Install dependencies:
```
  # apt install libzmq3-dev libgtk-3-dev libappindicator3-dev golang pkg-build
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
  $ brew install golang zmq pkg-build
  $ git clone 0xacab.org/leap/bitmask-systray
  $ cd bitmask-systray
  $ go get .
  $ go build
```

Run it
-------------
bitmask-systray assumes that you already have bitmaskd running.
