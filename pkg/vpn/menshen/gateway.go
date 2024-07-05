package menshen

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"0xacab.org/leap/bitmask-core/models"
	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"github.com/rs/zerolog/log"
)

const (
	maxGateways = 3 // how many OpenVPN remotes used for OpenVPN (--remote argument)
)

// Returns a gateway for a given ip address. Use case: We start OpenVPN with multiple
// --remote arguments, but in the end we don't know to which endpoint we are connected.
// Using the management interface, OpenVPN tells us the ip we are connected to
func (m *Menshen) GetGatewayByIP(ip string) (bonafide.Gateway, error) {
	for _, gw := range m.Gateways {
		if gw.IPAddr == ip {
			return *NewBonafideGateway(gw), nil
		}
	}
	return bonafide.Gateway{}, fmt.Errorf("Could not find a gateway with ip %s", ip)
}

// Returns a list of gateways that we will connect to. First checks if automatic gateway
// selection should be used.
func (m *Menshen) GetBestGateways(transport string) ([]bonafide.Gateway, error) {
	var location string
	var err error

	if m.IsManualLocation() {
		location = m.userChoice
	} else {
		location, err = m.GetBestLocation(transport)
		if err != nil {
			return []bonafide.Gateway{}, err
		}
	}

	log.Info().
		Str("transport", transport).
		Str("location", location).
		Bool("manualLocationSelection", m.IsManualLocation()).
		Msg("Finding best gateways to connect with")

	var gateways []*models.ModelsGateway
	gws, found := m.gwsByLocation[location]
	if !found {
		return []bonafide.Gateway{}, fmt.Errorf("Could not find a gateway for location %s", location)
	}

	if len(gws) > 1 {
		// use random gateways per location - otherwise all users use the same n gateways for each location
		// TODO: maybe menshen returns a random/sorted list?
		rand.Shuffle(len(gws), func(i, j int) { gws[i], gws[j] = gws[j], gws[i] })
		log.Debug().
			Str("location", location).
			Str("shuffledGateways", strings.Join(getGatewayNames(gws), " ")).
			Msg("Shuffled gateways for location")
	}

	for i, gw := range gws {
		// just use up to maxGateways gateways
		if i == maxGateways {
			break
		}
		gateways = append(gateways, gw)
	}

	gatewayNames := getGatewayNames(gateways)
	log.Debug().
		Str("location", location).
		Str("gateways", strings.Join(gatewayNames, " ")).
		Int("gatewayCount", len(gateways)).
		Int("maxGateways", maxGateways).
		Msg("Found best gateways for location")
	return NewBonafideGatewayArray(gateways), nil
}

// Just a helper for debugging output that returns a list of hostnames
// for a list of gateway objects
func getGatewayNames(gateways []*models.ModelsGateway) []string {
	var gwNames []string
	for _, gw := range gateways {
		gwNames = append(gwNames, gw.Host)
	}
	return gwNames
}

// Asks menshen for gateways. The gateways are stored in m.Gateways
// Currently, there is not CountryCode filtering
// The vars m.gwLocations and m.gwsByLocation are updated
func (m *Menshen) FetchAllGateways(transport string) error {
	log.Trace().Msg("Fetching gateways from menshen")

	// TODO: implement obfsv4 support (transport can have the value "any")
	if transport == "obfs4" {
		errors.New("obfs4 is not supported for v5 right now")
	}

	// reset if called multiple times
	m.gwLocations = []string{}
	m.gwsByLocation = make(map[string][]*models.ModelsGateway)

	var err error
	// TODO: send CountryCode
	m.Gateways, err = m.api.GetGateways(nil)
	if err != nil {
		return err
	}

	// TODO: gw.Port instead of gw.Ports
	for i, gw := range m.Gateways {
		log.Debug().
			Str("host", gw.Host).
			Int64("port", gw.Ports[0]).
			Str("ip", gw.IPAddr).
			Str("location", strings.Title(gw.Location)).
			Str("protocol", gw.Transport).
			Str("transport", gw.Type).
			Msg("Got gateway from API")

		// TODO: get rid of the strings.Title stuff if menshen supports gateway identifier
		if !slices.Contains(m.gwLocations, strings.Title(gw.Location)) {
			m.gwLocations = append(m.gwLocations, strings.Title(gw.Location))
		}
		m.gwsByLocation[strings.Title(gw.Location)] = append(m.gwsByLocation[strings.Title(gw.Location)], m.Gateways[i])
	}
	m.updateLocationQualityMap(transport)
	return nil
}

// Sets m.userChoice to a location if the user selects a location in the GUI
func (m *Menshen) SetManualGateway(location string) {
	if !slices.Contains(m.gwLocations, location) {
		log.Warn().
			Str("location", location).
			Msg("Could not set invalid location")
		return
	}

	log.Info().
		Str("location", location).
		Msg("Setting manual location")
	m.userChoice = location
}

// Sets m.userChoice to an empty string (auto-select best gateway/location)
func (m *Menshen) SetAutomaticGateway() {
	log.Debug().Msg("Setting remote gateway to automatic")
	m.userChoice = ""
}
