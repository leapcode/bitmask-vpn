package obfsvpn

import (
	"net"

	"github.com/xtaci/kcp-go/v5"
)

const (
	DefaultKCPSendWindowSize    int = 65535
	DefaultKCPReceiveWindowSize int = 65535
	DefaultKCPReadBuffer        int = 16 * 1024 * 1024
	DefaultKCPWriteBuffer       int = 16 * 1024 * 1024
)

type KCPConfig struct {
	Enabled           bool
	SendWindowSize    int
	ReceiveWindowSize int
	ReadBuffer        int
	WriteBuffer       int
}

func DefaultKCPConfig() *KCPConfig {
	return &KCPConfig{
		Enabled:           true,
		SendWindowSize:    DefaultKCPSendWindowSize,
		ReceiveWindowSize: DefaultKCPReceiveWindowSize,
		ReadBuffer:        DefaultKCPReadBuffer,
		WriteBuffer:       DefaultKCPWriteBuffer,
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
		// https://github.com/skywind3000/kcp/blob/master/README.en.md#protocol-configuration
		kcpSession.SetNoDelay(1, 10, 2, 1)
		return kcpSession, nil
	}

}
