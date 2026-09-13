package menshen

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"0xacab.org/leap/menshen/models"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
			service, err := m.GetService()
			if err != nil {
				return bonafide.Gateway{}, fmt.Errorf("failed to determine gateway location for gateway with ip %s", ip)
			}

			return *NewBonafideGateway(gw, service), nil
		}
	}
	return bonafide.Gateway{}, fmt.Errorf("could not find a gateway with ip %s", ip)
}

// Returns a list of gateways that we will connect to. First checks if automatic gateway
// selection should be used.
func (m *Menshen) GetBestGateways(transport string) ([]bonafide.Gateway, error) {
	switch transport {
	case "obfs4", "kcp", "quic":
		return m.getBestBridges(transport)
	}

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
		return []bonafide.Gateway{}, fmt.Errorf("could not find a gateway for location %s", location)
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
	return NewBonafideGatewayArray(gateways, m.service), nil
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

func getBridgeNames(bridges []*models.ModelsBridge) []string {
	names := make([]string, 0, len(bridges))
	for _, br := range bridges {
		names = append(names, br.Host)
	}
	return names
}

func (m *Menshen) getBestBridges(transport string) ([]bonafide.Gateway, error) {
	if len(m.Bridges) == 0 {
		log.Debug().Msg("Bridges not yet fetched, fetching on-demand")
		bridges, err := m.api.GetAllBridges(nil)
		if err != nil {
			return []bonafide.Gateway{}, fmt.Errorf("failed to fetch bridges: %w", err)
		}
		m.Bridges = bridges
	}

	service, err := m.GetService()
	if err != nil {
		return []bonafide.Gateway{}, fmt.Errorf("failed to get service for bridge conversion: %w", err)
	}

	var selected []*models.ModelsBridge
	for _, br := range m.Bridges {
		if !supportsTransport(br, transport) {
			continue
		}
		if m.IsManualLocation() && br.Location != m.userChoice {
			continue
		}
		selected = append(selected, br)
	}

	if len(selected) == 0 {
		log.Warn().
			Str("transport", transport).
			Str("location", m.userChoice).
			Msg("No bridges found for transport/location")
		return []bonafide.Gateway{}, fmt.Errorf("no bridges available for transport %s", transport)
	}

	if len(selected) > 1 {
		rand.Shuffle(len(selected), func(i, j int) { selected[i], selected[j] = selected[j], selected[i] })
		log.Debug().
			Str("transport", transport).
			Strs("shuffledBridges", getBridgeNames(selected)).
			Msg("Shuffled bridges")
	}

	if len(selected) > maxGateways {
		selected = selected[:maxGateways]
	}

	log.Info().
		Str("transport", transport).
		Strs("bridges", getBridgeNames(selected)).
		Int("bridgeCount", len(selected)).
		Msg("Selected best bridges for transport")

	return NewBonafideGatewayArrayFromBridges(selected, service), nil
}

// supportsTransport checks whether a bridge supports the requested
// obfuscation transport. In the v5 bridge model, Type is the bridge type
// (e.g. "obfs4") and Transport is the underlying protocol (tcp, kcp, quic).
func supportsTransport(br *models.ModelsBridge, transport string) bool {
	switch transport {
	case "obfs4":
		return br.Type == "obfs4"
	case "kcp":
		return br.Type == "obfs4" && br.Transport == "kcp"
	case "quic":
		return br.Type == "obfs4" && br.Transport == "quic"
	default:
		return false
	}
}

// Asks menshen for gateways and bridges. The gateways are stored in m.Gateways,
// the bridges in m.Bridges. Currently, there is not CountryCode filtering.
// The vars m.gwLocations and m.gwsByLocation are updated.
func (m *Menshen) FetchAllGateways(transport string) error {
	log.Trace().Msg("Fetching gateways and bridges from menshen")

	// reset if called multiple times
	m.gwsByLocation = make(map[string][]*models.ModelsGateway)

	var (
		gateways []*models.ModelsGateway
		bridges  []*models.ModelsBridge
		service  *models.ModelsEIPService
	)
	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		var err error
		gateways, err = m.api.GetGateways(nil)
		return err
	})
	g.Go(func() error {
		var err error
		bridges, err = m.api.GetAllBridges(nil)
		if err != nil {
			log.Warn().Err(err).Msg("Could not fetch bridges from menshen")
		}
		return nil
	})
	g.Go(func() error {
		var err error
		service, err = m.GetService()
		return err
	})
	if err := g.Wait(); err != nil {
		return err
	}
	m.Gateways = gateways
	m.Bridges = bridges
	m.service = service

	// TODO: gw.Port instead of gw.Ports
	for _, gw := range m.Gateways {
		log.Debug().
			Str("host", gw.Host).
			Int64("port", gw.Port).
			Str("ip", gw.IPAddr).
			Str("location", gw.Location).
			Str("protocol", gw.Transport).
			Str("transport", gw.Type).
			Msg("Got gateway from API")

		_, found := m.service.Locations[gw.Location]
		if !found {
			log.Warn().
				Str("gateway location label: ", gw.Location).
				Str("service locations: ", fmt.Sprintf("%+v", m.service.Locations)).
				Msg("Gateway contains label which is not available in service locations")
			continue
		}

		list, found := m.gwsByLocation[gw.Location]
		if !found {
			list = make([]*models.ModelsGateway, 0)
		}

		list = append(list, gw)
		m.gwsByLocation[gw.Location] = list
	}
	m.updateLocationQualityMap(transport)
	return nil
}

// Sets m.userChoice to a location if the user selects a location in the GUI
func (m *Menshen) SetManualGateway(location string) {
	service, err := m.GetService()
	if err != nil {
		log.Warn().
			Str("location", location).
			Msg("Could determine locations.")
		return
	}
	_, found := service.Locations[location]
	if !found {
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

func (m *Menshen) SupportsObfs4() bool {
	for _, br := range m.Bridges {
		if br.Type == "obfs4" {
			return true
		}
	}
	return false
}

func (m *Menshen) SupportsKCP() bool {
	for _, br := range m.Bridges {
		if br.Type == "obfs4" && br.Transport == "kcp" {
			return true
		}
	}
	return false
}

func (m *Menshen) SupportsQUIC() bool {
	for _, br := range m.Bridges {
		if br.Type == "obfs4" && br.Transport == "quic" {
			return true
		}
	}
	return false
}

func (m *Menshen) SupportsHopping() bool {
	return false
}
