# ObfsVPN

The `obfsvpn` module contains a Go package that provides server and client components to 
use variants of the obfs4 obfuscation protocol. It is intended to be used as a
drop-in Pluggable Transport for OpenVPN connections (although it can be used
for other, more generic purposes).

A docker container will be provided to facilitate startng an OpenVPN service that
is accessible via the obfuscated proxy too.

## Protocol stack

```
--------------------
 application data
--------------------
      OpenVPN
--------------------
   obfsvpn proxy
--------------------
       obfs4
--------------------
   wire transport
--------------------
```

* Application data is written to the specified interface (typically a `tun`
  device started by `OpenVPN`).
* `OpenVPN` provides end-to-end encryption and a reliability layer. We'll be
  testing with the `2.5.x` branch of the reference OpenVPN implementation.
* `obfs4` is used for an extra layer of encryption and obfuscation. It is a
  look-like-nothing protocol that also hides the key exchange to the eyes of
  the censor.
* Wire transport is, by default, TCP. Other transports will be explored to
  facilitate evasion: `KCP`, `QUIC`?

## Testing

...

## Android

Assuming you have the `android ndk` in place, you can build the bindings for android using `gomobile`:

```
make build-android
```

