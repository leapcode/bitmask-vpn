package obfsvpn

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"

	"github.com/xtaci/kcp-go/v5"
	"gitlab.com/yawning/obfs4.git/common/ntor"
	"gitlab.com/yawning/obfs4.git/transports/base"
	"gitlab.com/yawning/obfs4.git/transports/obfs4"
	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
)

// ListenConfig contains options for listening to an address.
// If Seed is not set it defaults to a randomized value.
// If StateDir is not set the current working directory is used.
type ListenConfig struct {
	ListenConfig net.ListenConfig
	KCPConfig    KCPConfig

	NodeID     *ntor.NodeID
	PrivateKey *ntor.PrivateKey
	PublicKey  string
	Seed       [ntor.KeySeedLength]byte
	StateDir   string
}

// NewListenConfig returns a ListenConfig and any error during the initialization.
// perhaps this is redundant, but using the same json format than ss for debug.
func NewListenConfig(nodeIDStr, privKeyStr, pubKeyStr, seedStr, stateDir string, kcpConfig KCPConfig) (*ListenConfig, error) {
	var err error
	var seed [ntor.KeySeedLength]byte
	var nodeID *ntor.NodeID
	private := new(ntor.PrivateKey)

	if nodeID, err = ntor.NodeIDFromHex(nodeIDStr); err != nil {
		return nil, err
	}

	raw, err := hex.DecodeString(privKeyStr)
	if err != nil {
		return nil, err
	}
	copy(private[:], raw)

	s, err := hex.DecodeString(seedStr)
	if err != nil {
		return nil, err
	}
	copy(seed[:], s)

	lc := &ListenConfig{
		NodeID:     nodeID,
		PrivateKey: private,
		PublicKey:  pubKeyStr,
		Seed:       seed,
		StateDir:   stateDir,
		KCPConfig:  kcpConfig,
	}
	return lc, nil
}

// NewListenConfigCert creates a listener config by unpacking the node ID from
// its certificate.
// The private key must still be specified.
func NewListenConfigCert(cert string) (*ListenConfig, error) {
	nodeID, _, err := unpackCert(cert)
	if err != nil {
		return nil, err
	}
	return &ListenConfig{
		NodeID: nodeID,
	}, nil
}

// Wrap takes an existing net.Listener and wraps it in a listener that is
// configured to perform the ntor handshake and copy data through the obfuscated conn.
// Values from the inner net.ListenConfig are ignored.
func (lc *ListenConfig) Wrap(ctx context.Context, ln net.Listener) (*Listener, error) {
	args := make(pt.Args)
	args.Add("node-id", lc.NodeID.Hex())
	args.Add("private-key", lc.PrivateKey.Hex())
	seed := ntor.KeySeed{}
	if bytes.Equal(lc.Seed[:], seed[:]) {
		_, err := rand.Read(seed[:])
		if err != nil {
			return nil, err
		}
	} else {
		seed = lc.Seed
	}

	args.Add("drbg-seed", hex.EncodeToString(seed[:]))
	args.Add("public-key", lc.PublicKey)
	sf, err := (&obfs4.Transport{}).ServerFactory(lc.StateDir, &args)
	if err != nil {
		return nil, err
	}
	return &Listener{sf: sf, ln: ln}, nil
}

// NewServerState will create all the state (node-id, private-key, public-key,
// drbg-seed, iat-mode) that the OBFS4 server needs to function. It will write
// it into the passed stateDir folder.
func NewServerState(stateDir string) error {
	args := make(pt.Args)
	_, err := (&obfs4.Transport{}).ServerFactory(stateDir, &args)
	if err != nil {
		return err
	}
	return err
}

// Listen listens on the local network address.
// See func net.Dial for a description of the network and address parameters.
func (lc *ListenConfig) Listen(ctx context.Context, address string) (*Listener, error) {
	if lc.KCPConfig.Enabled {
		log.Println("kcp listen on", address)
		ln, err := kcp.ListenWithOptions(address, nil, 10, 3)
		if err != nil {
			return nil, err
		}

		if err := ln.SetReadBuffer(lc.KCPConfig.ReadBuffer); err != nil {
			return nil, err
		}
		if err := ln.SetWriteBuffer(lc.KCPConfig.WriteBuffer); err != nil {
			return nil, err
		}

		wrappedListener, err := lc.Wrap(ctx, ln)
		if err != nil {
			return nil, err
		}
		wrappedListener.kcpConfig = lc.KCPConfig
		return wrappedListener, nil
	}
	ln, err := lc.ListenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	return lc.Wrap(ctx, ln)
}

// Listener is a network listener that accepts obfuscated connections and
// performs the ntor handshake on them.
type Listener struct {
	sf        base.ServerFactory
	ln        net.Listener
	kcpConfig KCPConfig
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.ln.Accept()
	if err != nil {
		return nil, err
	}
	kcpSession, ok := conn.(*kcp.UDPSession)
	if ok {
		kcpSession.SetStreamMode(true)
		kcpSession.SetWindowSize(l.kcpConfig.SendWindowSize, l.kcpConfig.ReceiveWindowSize)
		// https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
		nd := 0
		if l.kcpConfig.NoDelay {
			nd = 1
		}
		nc := 0
		if l.kcpConfig.DisableFlowControl {
			nc = 1
		}
		kcpSession.SetNoDelay(nd, l.kcpConfig.Interval, l.kcpConfig.Resend, nc)
		kcpSession.SetMtu(l.kcpConfig.MTU)
	}
	conn, err = l.sf.WrapConn(conn)
	return conn, err
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *Listener) Close() error {
	return l.ln.Close()
}

// Addr returns the listener's network address.
func (l *Listener) Addr() net.Addr {
	return l.ln.Addr()
}
