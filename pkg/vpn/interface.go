package vpn

import (
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
)

type apiInterface interface {
	NeedsCredentials() bool
	DoLogin(username, password string) (bool, error)
	GetLocationQualityMap(transport string) map[string]float64
	GetLocationLabels(transport string) map[string][]string
	SetManualGateway(label string)
	SetAutomaticGateway()
	IsManualLocation() bool
	IsUDPAvailable() bool
	GetBestLocation(transport string) (string, error)
	GetPemCertificate() ([]byte, error)
	GetOpenvpnArgs() ([]string, error)
	GetGatewayByIP(ip string) (bonafide.Gateway, error)
	GetBestGateways(transport string) ([]bonafide.Gateway, error)
	FetchAllGateways(transport string) error
	GetSnowflakeCh() chan *snowflake.StatusEvent
}
