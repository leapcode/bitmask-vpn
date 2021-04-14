# obfs4


This is a look-like nothing obfuscation protocol that incorporates ideas and concepts from Philipp Winter's ScrambleSuit protocol. The obfs naming was chosen primarily because it was shorter, in terms of protocol ancestery obfs4 is much closer to ScrambleSuit than obfs2/obfs3.

The notable differences between ScrambleSuit and obfs4:

* The handshake always does a full key exchange (no such thing as a Session Ticket Handshake).
* The handshake uses the Tor Project's ntor handshake with public keys obfuscated via the Elligator 2 mapping.
* The link layer encryption uses NaCl secret boxes (Poly1305/XSalsa20).
* As an added bonus, obfs4proxy also supports acting as an obfs2/3 client and bridge to ease the transition to the new protocol.

## Using obfs4


### Go Version:

obfs4 is one of the transports available in the [Shapeshifter-Transports library](https://github.com/OperatorFoundation/Shapeshifter-Transports).

1. First, you need to create a dialer
    `dialer := proxy.Direct`
    
2. Create an instance of an obfs4 server
    `obfs4Transport := obfs4.Transport{
     		CertString: "InsertCertStringHere",
     		IatMode:    0 or 1,
     		Address:    "InsertAddressHere",
     		Dialer:     dialer,}`

5. Call Dial on obfs4Transport:
    `_, err := obfs4Transport.Dial()`
