ShapeShifter library
--------------------

Heavily based on the shapeshifter-dispatcher:
https://github.com/OperatorFoundation/shapeshifter-dispatcher/


To use it:
```go
	ss := ShapeShifter{
		Cert:      "cert",
		Target:    "ip:port",
		SocksAddr: "127.0.0.1:4430",
	}
	err := ss.Open()
	if err != nil {
		return err
	}
	defer ss.Close()
```

And now you can tunnel your protocol into `127.0.0.1:4430`.
