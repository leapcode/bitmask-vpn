package obfsvpn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"

	//pt "git.torproject.org/pluggable-transports/goptlib.git"
	"gitlab.com/yawning/obfs4.git/common/ntor"
	"gitlab.com/yawning/obfs4.git/transports/base"
	"gitlab.com/yawning/obfs4.git/transports/obfs4"
	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
)

const (
	ptArgNode = "node-id"
	ptArgKey  = "public-key"
	ptArgMode = "iat-mode"
	ptArgCert = "cert"
)

const (
	certLength = ntor.NodeIDLength + ntor.PublicKeyLength
)

var (
	ErrCannotDial = errors.New("cannot dial")
)

// IATMode determines the amount of time sent between packets.
type IATMode int

// Valid IAT modes.
const (
	IATNone IATMode = iota
	IATEnabled
	IATParanoid
)

// Dialer contains options for connecting to an address and obfuscating traffic
// with the obfs4 protocol.
// It performs the ntor handshake on all dialed connections.
type Dialer struct {
	Dialer net.Dialer

	NodeID    *ntor.NodeID
	PublicKey *ntor.PublicKey
	IATMode   IATMode
	DialFunc  func(string, string) (net.Conn, error)

	ptArgs        pt.Args
	clientFactory base.ClientFactory
}

func packCert(node *ntor.NodeID, public *ntor.PublicKey) string {
	cert := make([]byte, 0, certLength)
	cert = append(cert, node[:]...)
	cert = append(cert, public[:]...)

	return base64.RawStdEncoding.EncodeToString(cert)
}

func unpackCert(cert string) (*ntor.NodeID, *ntor.PublicKey, error) {
	if l := base64.RawStdEncoding.DecodedLen(len(cert)); l != certLength {
		return nil, nil, fmt.Errorf("cert length %d is invalid", l)
	}
	decoded, err := base64.RawStdEncoding.DecodeString(cert)
	if err != nil {
		return nil, nil, err
	}

	nodeID, err := ntor.NewNodeID(decoded[:ntor.NodeIDLength])
	if err != nil {
		return nil, nil, err
	}
	pubKey, err := ntor.NewPublicKey(decoded[ntor.NodeIDLength:])
	if err != nil {
		return nil, nil, err
	}
	return nodeID, pubKey, nil
}

// NewDialerFromCert creates a dialer from a node certificate.
func NewDialerFromCert(cert string) (*Dialer, error) {
	nodeID, publicKey, err := unpackCert(cert)

	if err != nil {
		return nil, err
	}
	d := &Dialer{
		NodeID:    nodeID,
		PublicKey: publicKey,
	}
	return d, nil
}

// NewDialerFromArgs creates a dialer from existing pluggable transport arguments.
func NewDialerFromArgs(args pt.Args) (*Dialer, error) {
	clientFactory, err := (&obfs4.Transport{}).ClientFactory("")
	if err != nil {
		return nil, err
	}
	nodeHex, _ := args.Get(ptArgNode)
	node, err := ntor.NodeIDFromHex(nodeHex)
	if err != nil {
		return nil, err
	}
	keyHex, _ := args.Get(ptArgKey)
	pub, err := ntor.PublicKeyFromHex(keyHex)
	if err != nil {
		return nil, err
	}
	iatModeStr, _ := args.Get(ptArgMode)
	iatMode, err := strconv.Atoi(iatModeStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing IAT mode to integer: %w", err)
	}
	return &Dialer{
		NodeID:    node,
		PublicKey: pub,
		IATMode:   IATMode(iatMode),

		ptArgs:        args,
		clientFactory: clientFactory,
	}, nil
}

// Dial creates an outbound net.Conn and performs the ntor handshake.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	ctx := context.Background()
	return d.dial(ctx, network, address, func(network, address string) (net.Conn, error) {
		conn, err := d.Dialer.DialContext(ctx, network, address)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCannotDial, err)
		}
		return conn.(*net.TCPConn), err
	})
}

// DialContext creates an outbound net.Conn and performs the ntor handshake.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dial(ctx, network, address, func(network, address string) (net.Conn, error) {
		return d.Dialer.DialContext(ctx, network, address)
	})
}

func (d *Dialer) dial(ctx context.Context, network, address string, f func(network, address string) (net.Conn, error)) (net.Conn, error) {
	if d.clientFactory == nil {
		clientFactory, err := (&obfs4.Transport{}).ClientFactory("")
		if err != nil {
			return nil, err
		}
		d.clientFactory = clientFactory
	}
	ptArgs := d.Args()
	args, err := d.clientFactory.ParseArgs(&ptArgs)
	if err != nil {
		return nil, err
	}
	if d.DialFunc != nil {
		f = d.DialFunc
	}
	return d.clientFactory.Dial(network, address, f, args)
}

// Args returns the dialers options as pluggable transport arguments.
// The args include valid args for the "new" (version >= 0.0.3) bridge lines
// that use a unified "cert" argument as well as the legacy lines that use a
// separate Node ID and Public Key.
func (d *Dialer) Args() pt.Args {
	if d.ptArgs == nil {
		d.ptArgs = make(pt.Args)
		d.ptArgs.Add(ptArgNode, d.NodeID.Hex())
		d.ptArgs.Add(ptArgKey, d.PublicKey.Hex())
		d.ptArgs.Add(ptArgMode, strconv.Itoa(int(d.IATMode)))
		d.ptArgs.Add(ptArgCert, packCert(d.NodeID, d.PublicKey))
	}
	return d.ptArgs
}
