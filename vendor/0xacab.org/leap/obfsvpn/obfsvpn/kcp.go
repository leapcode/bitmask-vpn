package obfsvpn

import (
	"net"

	"github.com/xtaci/kcp-go/v5"
)

const (
	DefaultKCPSendWindowSize    int  = 65535
	DefaultKCPReceiveWindowSize int  = 65535
	DefaultKCPReadBuffer        int  = 16 * 1024 * 1024
	DefaultKCPWriteBuffer       int  = 16 * 1024 * 1024
	DefaultNoDelay              bool = true
	DefaultInterval             int  = 10
	DefaultResend               int  = 2
	DefaultDisableFlowControl   bool = true
	DefaultMTU                  int  = 1400
)

// https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
type KCPConfig struct {
	Enabled            bool
	SendWindowSize     int
	ReceiveWindowSize  int
	ReadBuffer         int
	WriteBuffer        int
	NoDelay            bool
	Interval           int
	Resend             int
	DisableFlowControl bool
	MTU                int
}

func DefaultKCPConfig() *KCPConfig {
	return &KCPConfig{
		Enabled:            true,
		SendWindowSize:     DefaultKCPSendWindowSize,
		ReceiveWindowSize:  DefaultKCPReceiveWindowSize,
		ReadBuffer:         DefaultKCPReadBuffer,
		WriteBuffer:        DefaultKCPWriteBuffer,
		NoDelay:            DefaultNoDelay,
		Interval:           DefaultInterval,
		Resend:             DefaultResend,
		DisableFlowControl: DefaultDisableFlowControl,
		MTU:                DefaultMTU,
	}
}

func GetKCPDialer(kcpConfig KCPConfig, logger func(format string, a ...interface{})) func(network, address string) (net.Conn, error) {
	return func(network, address string) (net.Conn, error) {
		if logger != nil {
			logger("Dialing kcp://%s", address)
		}
		kcpSession, err := kcp.DialWithOptions(address, nil, 10, 3)
		if err != nil {
			return nil, err
		}
		kcpSession.SetStreamMode(true)
		kcpSession.SetWindowSize(kcpConfig.SendWindowSize, kcpConfig.ReceiveWindowSize)
		if err := kcpSession.SetReadBuffer(kcpConfig.ReadBuffer); err != nil {
			return nil, err
		}
		if err := kcpSession.SetWriteBuffer(kcpConfig.WriteBuffer); err != nil {
			return nil, err
		}
		nd := 0
		if kcpConfig.NoDelay {
			nd = 1
		}
		nc := 0
		if kcpConfig.DisableFlowControl {
			nc = 1
		}
		kcpSession.SetNoDelay(nd, kcpConfig.Interval, kcpConfig.Resend, nc)
		kcpSession.SetMtu(kcpConfig.MTU)
		return kcpSession, nil
	}

}

func GetKCPStats() *kcp.Snmp {
	return kcp.DefaultSnmp.Copy()
}
