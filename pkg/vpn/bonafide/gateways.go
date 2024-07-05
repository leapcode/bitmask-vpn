package bonafide

import (
	"errors"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	maxGateways = 3
)

// Load reflects the fullness metric that menshen returns, if available.
type Load struct {
	gateway  *Gateway
	Fullness float64
	Overload bool
}

// A Gateway is a representation of gateways that is independent of the api version.
// If a given physical location offers different transports, they will appear
// as separate gateways, so make sure to filter them.
type Gateway struct {
	Host         string
	IPAddress    string
	Location     string
	LocationName string
	CountryCode  string
	Ports        []string
	Protocols    []string
	Options      map[string]string
	Transport    string
}

/* gatewayDistance is used in the timezone distance fallback */
type gatewayDistance struct {
	gateway  Gateway
	distance int
}

type gatewayPool struct {
	/* available is the unordered list of gateways from eip-service, we use if as source-of-truth for now. */
	available  []Gateway
	userChoice string

	/* byLocation is a map from location to an array of hostnames */
	byLocation map[string][]*Gateway

	/* recommended is an array of hostnames, fetched from the old geoip service. */
	recommended []Load

	/* TODO locations are just used to get the timezone for each gateway. I
	* think it's easier to just merge that info into the version-agnostic
	* Gateway, that is passed from the eipService, and do not worry with
	* the location here */
	locations map[string]Location
}

func (gw Gateway) isTransport(transport string) bool {
	return transport == "any" || gw.Transport == transport
}

func (p *gatewayPool) populateLocationList() {
	for i, gw := range p.available {
		p.byLocation[gw.Location] = append(p.byLocation[gw.Location], &p.available[i])
	}
}

func (p *gatewayPool) getLocations() []string {
	c := make([]string, 0)
	if p == nil || p.byLocation == nil || len(p.byLocation) == 0 {
		return c
	}
	if len(p.byLocation) != 0 {
		for city := range p.byLocation {
			c = append(c, city)
		}
	}
	return c
}

func (p *gatewayPool) isValidLocation(location string) bool {
	locations := p.getLocations()
	valid := stringInSlice(location, locations)
	return valid
}

/* returns a map of location: fullness for the ui to use */
func (p *gatewayPool) getLocationQualityMap(transport string) map[string]float64 {
	locations := p.getLocations()
	cm := make(map[string]float64)
	if len(locations) == 0 {
		return cm
	}
	if len(p.recommended) != 0 {
		for idx, gw := range p.recommended {
			if gw.gateway.Transport != transport {
				continue
			}
			if _, ok := cm[gw.gateway.Location]; ok {
				continue
			}
			if gw.Fullness != -1 {
				cm[gw.gateway.Location] = gw.Fullness
			} else {
				cm[gw.gateway.Location] = 1 - float64(idx)/float64(len(p.recommended))
			}
		}
	} else {
		for _, location := range locations {
			cm[location] = -1
		}
	}
	return cm
}

/* returns a map of location: labels for the ui to use */
func (p *gatewayPool) getLocationLabels(transport string) map[string][]string {
	cm := make(map[string][]string)
	locations := p.getLocations()
	if len(locations) == 0 {
		return cm
	}
	for _, loc := range locations {
		current := p.locations[loc]
		cm[loc] = []string{current.Name, current.CountryCode}
	}
	return cm
}

/* this method should only be used if we have no usable menshen list. */
func (p *gatewayPool) getRandomGatewaysByLocation(location, transport string) ([]Gateway, error) {
	if !p.isValidLocation(location) {
		return []Gateway{}, errors.New("bonafide: BUG not a valid location: " + location)
	}
	gws := p.byLocation[location]
	if len(gws) == 0 {
		return []Gateway{}, errors.New("bonafide: BUG no gw for location: " + location)
	}

	r := rand.New(rand.NewSource(time.Now().Unix()))
	r.Shuffle(len(gws), func(i, j int) { gws[i], gws[j] = gws[j], gws[i] })

	var gateways []Gateway
	for _, gw := range gws {
		if gw.isTransport(transport) {
			gateways = append(gateways, *gw)
		}
		if len(gateways) == maxGateways {
			break
		}
	}
	if len(gateways) == 0 {
		return []Gateway{}, errors.New("bonafide: BUG could not find any gateway for that location")
	}

	return gateways, nil
}

func (p *gatewayPool) getGatewaysFromMenshenByLocation(location, transport string) ([]Gateway, error) {
	if !p.isValidLocation(location) {
		return []Gateway{}, errors.New("bonafide: BUG not a valid location: " + location)
	}
	gws := p.byLocation[location]
	if len(gws) == 0 {
		return []Gateway{}, errors.New("bonafide: BUG no gw for location: " + location)
	}

	var gateways []Gateway
	for _, gw := range p.recommended {
		if !gw.gateway.isTransport(transport) {
			continue
		}
		for _, locatedGw := range gws {
			if locatedGw.Host == gw.gateway.Host {
				gateways = append(gateways, *locatedGw)
				break
			}
		}
		if len(gateways) == maxGateways {
			break
		}
	}
	if len(gateways) == 0 {
		return []Gateway{}, errors.New("bonafide: BUG could not find any gateway for that location")
	}

	return gateways, nil
}

/* used when we select a hostname in the ui and we want to know the gateway details */
func (p *gatewayPool) getGatewayByHost(host string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.Host == host {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid host name")
}

/* used when we want to know gateway details after we know what IP openvpn has connected to */
func (p *gatewayPool) getGatewayByIP(ip string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.IPAddress == ip {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid ip address")
}

/* this perhaps could be made more explicit */
func (p *gatewayPool) setAutomaticChoice() {
	p.userChoice = ""
}

/* set a user manual override for gateway location */
func (p *gatewayPool) setUserChoice(location string) error {
	if !p.isValidLocation(location) {
		return errors.New("bonafide: not a valid city for gateway choice")
	}
	p.userChoice = location
	return nil
}

func (p *gatewayPool) isManualLocation() bool {
	return len(p.userChoice) != 0
}

/* set the recommended field from an ordered array. needs to be modified if menshen passed an array of Loads */
func (p *gatewayPool) setRecommendedGateways(geo *geoLocation) {
	var recommended []Load
	if len(geo.SortedGateways) != 0 {
		for _, gw := range geo.SortedGateways {
			found := false
			for i := range p.available {
				if p.available[i].Host == gw.Host {
					recommendedGw := Load{
						Fullness: gw.Fullness,
						Overload: gw.Overload,
						gateway:  &p.available[i],
					}
					recommended = append(recommended, recommendedGw)
					found = true
				}
			}
			if !found {
				log.Warn().
					Str("host", gw.Host).
					Msg("Invalid host in recommended list of hostnames")
				return
			}
		}
	} else {
		// If there is not sorted gatways, it means that the old menshen API is being used
		// let's use the list of hosts then
		for _, host := range geo.Gateways {
			found := false
			for i := range p.available {
				if p.available[i].Host == host {
					recommendedGw := Load{
						Fullness: -1,
						gateway:  &p.available[i],
					}
					recommended = append(recommended, recommendedGw)
					found = true
				}
			}
			if !found {
				log.Warn().
					Str("host", host).
					Msg("Invalid host in recommended list of hostnames")
				return
			}
		}
	}

	p.recommended = recommended
}

/* get at most max gateways. the method of picking depends on whether we're
* doing manual override, and if we got useful info from menshen */
func (p *gatewayPool) getBest(transport string, tz, max int) ([]Gateway, error) {
	if hostname := os.Getenv("LEAP_GW"); hostname != "" {
		log.Debug().
			Str("hostname", hostname).
			Msg("Gateway selection manually overriden")
		return p.getGatewaysByHostname(hostname)
	}
	if p.isManualLocation() {
		if len(p.recommended) != 0 {
			return p.getGatewaysFromMenshenByLocation(p.userChoice, transport)
		} else {
			return p.getRandomGatewaysByLocation(p.userChoice, transport)
		}
	} else if len(p.recommended) != 0 {
		return p.getGatewaysFromMenshen(transport, max)
	} else {
		return p.getGatewaysByTimezone(transport, tz, max)
	}
}

/* returns the location for the first recommended gateway */
func (p *gatewayPool) getBestLocation(transport string, tz int) string {
	best, err := p.getBest(transport, tz, 1)
	if err != nil {
		return ""
	}
	if len(best) != 1 {
		return ""
	}
	return best[0].Location

}

func (p *gatewayPool) getAll(transport string, tz int) error {
	if (&gatewayPool{} == p) {
		log.Warn().Msg("getAll tried to access uninitialized struct")
		return nil
	}

	log.Debug().Msg("seems to be initialized...")
	_, err := p.getGatewaysByTimezone(transport, tz, 999)
	return err
}

/* picks at most max gateways, filtering by transport, from the ordered list menshen returned */
func (p *gatewayPool) getGatewaysFromMenshen(transport string, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	for _, gw := range p.recommended {
		if !gw.gateway.isTransport(transport) {
			continue
		}
		gws = append(gws, *gw.gateway)
		if len(gws) == max {
			break
		}
	}
	return gws, nil
}

/* the old timezone based heuristic, when everything goes wrong */
func (p *gatewayPool) getGatewaysByTimezone(transport string, tzOffsetHours, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	gwVector := []gatewayDistance{}

	for _, gw := range p.available {
		if !gw.isTransport(transport) {
			continue
		}
		distance := 13
		gwOffset, err := strconv.Atoi(p.locations[gw.Location].Timezone)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("Could not sort gateways")
			return gws, err
		}
		distance = tzDistance(tzOffsetHours, gwOffset)
		gwVector = append(gwVector, gatewayDistance{gw, distance})
	}
	rand.Seed(time.Now().UnixNano())
	cmp := func(i, j int) bool {
		if gwVector[i].distance == gwVector[j].distance {
			return rand.Intn(2) == 1
		}
		return gwVector[i].distance < gwVector[j].distance
	}
	sort.Slice(gwVector, cmp)

	for _, gw := range gwVector {
		gws = append(gws, gw.gateway)
		if len(gws) == max {
			break
		}
	}
	return gws, nil
}

// getGatewaysByHostname filters the gateway pool by hostname. If it finds a
// gateway matching the passed hostname, it will return a Gateway array with
// exactly one gateway. It will also return an error (which is always nil at
// the moment, but for coherence with similar methods).
func (p *gatewayPool) getGatewaysByHostname(hostname string) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	for _, gw := range p.available {
		if gw.Host == hostname {
			gws = append(gws, gw)
		}
	}
	return gws, nil
}

func newGatewayPool(eip *eipService) *gatewayPool {
	p := gatewayPool{}
	p.available = eip.getGateways()
	p.locations = eip.Locations
	p.byLocation = make(map[string][]*Gateway)
	p.populateLocationList()
	return &p
}

func tzDistance(offset1, offset2 int) int {
	abs := func(x int) int {
		if x < 0 {
			return -x
		}
		return x
	}
	distance := abs(offset1 - offset2)
	if distance > 12 {
		distance = 24 - distance
	}
	return distance
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
