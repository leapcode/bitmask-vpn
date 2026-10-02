package vpn

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	obfsvpnClient "0xacab.org/leap/obfsvpn/client"
)

func TestStopObfsvpnProxy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, err := obfsvpnClient.NewClient(ctx, cancel, obfsvpnClient.Config{
		ProxyAddr:  "127.0.0.1:8080",
		RemoteIP:   "127.0.0.1",
		RemotePort: "12345",
	})
	require.NoError(t, err, "Could not create obfsvpn client")

	b := Bitmask{obfsvpnProxy: client}
	b.stopObfsvpnProxy()
	require.Nil(t, b.obfsvpnProxy, "proxy reference not cleared after stop")

	// stopping again must be a safe no-op
	b.stopObfsvpnProxy()
	require.Nil(t, b.obfsvpnProxy, "double stop should not resurrect the proxy")
}
