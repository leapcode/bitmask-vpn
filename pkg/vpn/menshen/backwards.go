package menshen

import (
	"fmt"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/menshen/models"
)

func NewBonafideGatewayArray(gatewaysV5 []*models.ModelsGateway) []bonafide.Gateway {
	gws := make([]bonafide.Gateway, 0)
	for _, gw := range gatewaysV5 {
		transitGateway := NewBonafideGateway(gw)
		gws = append(gws, *transitGateway)
	}
	return gws
}

func NewBonafideGateway(v5Gateway *models.ModelsGateway) *bonafide.Gateway {
	transitGateway := &bonafide.Gateway{
		Host:         v5Gateway.Host,
		IPAddress:    v5Gateway.IPAddr,
		Location:     v5Gateway.Location,
		LocationName: strings.Title(v5Gateway.Location),
		CountryCode:  getCountryCodeForLocation(v5Gateway.Location),
		Ports:        []string{fmt.Sprintf("%d", v5Gateway.Port)},
		Protocols:    []string{v5Gateway.Type},
		//Options:      v5Gateway.Options,
		//Transport:    v5Gateway.Transport,
	}
	return transitGateway
}
