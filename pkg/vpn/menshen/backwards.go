package menshen

import (
	"fmt"
	"strconv"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/menshen/models"
)

func NewBonafideGatewayArray(gatewaysV5 []*models.ModelsGateway, service *models.ModelsEIPService) []bonafide.Gateway {
	gws := make([]bonafide.Gateway, 0)
	for _, gw := range gatewaysV5 {
		transitGateway := NewBonafideGateway(gw, service)
		gws = append(gws, *transitGateway)
	}
	return gws
}

func NewBonafideGateway(v5Gateway *models.ModelsGateway, service *models.ModelsEIPService) *bonafide.Gateway {
	transitGateway := &bonafide.Gateway{
		Host:         v5Gateway.Host,
		IPAddress:    v5Gateway.IPAddr,
		Location:     v5Gateway.Location,
		LocationName: getLocationName(v5Gateway.Location, service),
		CountryCode:  getCountryCodeForLocation(v5Gateway.Location, service),
		Ports:        []string{fmt.Sprintf("%d", v5Gateway.Port)},
		Protocols:    []string{v5Gateway.Type},
		//Options:      v5Gateway.Options,
		//Transport:    v5Gateway.Transport,
	}
	return transitGateway
}

func NewBonafideGatewayFromBridge(bridge *models.ModelsBridge, service *models.ModelsEIPService) *bonafide.Gateway {
	options := make(map[string]string)
	for k, v := range bridge.Options {
		options[k] = fmt.Sprintf("%v", v)
	}

	return &bonafide.Gateway{
		Host:         bridge.Host,
		IPAddress:    bridge.IPAddr,
		Location:     bridge.Location,
		LocationName: getLocationName(bridge.Location, service),
		CountryCode:  getCountryCodeForLocation(bridge.Location, service),
		Ports:        []string{strconv.FormatInt(bridge.Port, 10)},
		Protocols:    []string{bridge.Transport},
		Options:      options,
		Transport:    bridge.Type,
	}
}

func NewBonafideGatewayArrayFromBridges(bridges []*models.ModelsBridge, service *models.ModelsEIPService) []bonafide.Gateway {
	gws := []bonafide.Gateway{}
	for _, br := range bridges {
		gws = append(gws, *NewBonafideGatewayFromBridge(br, service))
	}
	return gws
}
