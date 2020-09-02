package bonafide

import (
	"errors"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

const (
	maxGateways = 3
)

// A Gateway is a representation of gateways that is independent of the api version.
// If a given physical location offers different transports, they will appear as separate gateways.
type Gateway struct {
	Host      string
	IPAddress string
	Location  string
	Ports     []string
	Protocols []string
	Options   map[string]string
	Transport string
	Label     string
}

/* TODO add a String method with a human representation: Label (cc) */
/* For that, we should pass the locations to genLabels, and generate a string repr */

type gatewayDistance struct {
	gateway  Gateway
	distance int
}

type gatewayPool struct {
	available []Gateway
	/* ranked is, for now, just an array of hostnames (fetched from the
	geoip service). it should be a map in the future, to keep track of
	quantitative metrics */
	ranked     []string
	userChoice string
	locations  map[string]location
}

/* genLabels generates unique, human-readable labels for a gateway. It gives a serial
   number to each gateway in the same location (paris-1, paris-2,...). The
   current implementation will give a different label to each transport.
*/
func (p *gatewayPool) genLabels() {
	acc := make(map[string]int)
	for i, gw := range p.available {
		if _, count := acc[gw.Location]; !count {
			acc[gw.Location] = 1
		} else {
			acc[gw.Location] += 1
		}
		gw.Label = gw.Location + "-" + strconv.Itoa(acc[gw.Location])
		p.available[i] = gw
	}
	/* skip suffix if only one occurence */
	for i, gw := range p.available {
		if acc[gw.Location] == 1 {
			gw.Label = gw.Location
			p.available[i] = gw
		}
	}
}

func (p *gatewayPool) getLabels() []string {
	labels := make([]string, 0)
	for _, gw := range p.available {
		labels = append(labels, gw.Label)
	}
	/* TODO return error if called when no labels have been generated */
	return labels
}

func (p *gatewayPool) isValidLabel(label string) bool {
	labels := p.getLabels()
	valid := stringInSlice(label, labels)
	return valid
}

func (p *gatewayPool) getGatewayByLabel(label string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.Label == label {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid label")
}

func (p *gatewayPool) getGatewayByIP(ip string) (Gateway, error) {
	for _, gw := range p.available {
		if gw.IPAddress == ip {
			return gw, nil
		}
	}
	return Gateway{}, errors.New("bonafide: not a valid ip address")
}

func (p *gatewayPool) setAutomaticChoice() {
	p.userChoice = ""
}

func (p *gatewayPool) setUserChoice(label string) error {
	if !p.isValidLabel(label) {
		return errors.New("bonafide: not a valid label for gateway choice")
	}
	p.userChoice = label
	return nil
}

func (p *gatewayPool) setRanking(hostnames []string) {
	hosts := make([]string, 0)
	for _, gw := range p.available {
		hosts = append(hosts, gw.Host)
	}

	for _, host := range hostnames {
		if !stringInSlice(host, hosts) {
			log.Println("ERROR: invalid host in ranked hostnames", host)
			return
		}
	}

	p.ranked = hostnames
}

func (p *gatewayPool) getBest(transport string, tz, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	if len(p.userChoice) != 0 {
		gw, err := p.getGatewayByLabel(p.userChoice)
		gws = append(gws, gw)
		return gws, err
	} else if len(p.ranked) != 0 {
		return p.getGatewaysByServiceRank(transport, max)
	} else {
		return p.getGatewaysByTimezone(transport, tz, max)
	}
}

func (p *gatewayPool) getAll(transport string, tz int) ([]Gateway, error) {
	if len(p.ranked) != 0 {
		return p.getGatewaysByServiceRank(transport, 999)
	} else {
		return p.getGatewaysByTimezone(transport, tz, 999)
	}
}

func (p *gatewayPool) getGatewaysByServiceRank(transport string, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	for _, host := range p.ranked {
		for _, gw := range p.available {
			if gw.Transport != transport {
				continue
			}
			if gw.Host == host {
				gws = append(gws, gw)
			}
			if len(gws) == max {
				goto end
			}
		}
	}
end:
	return gws, nil
}

func (p *gatewayPool) getGatewaysByTimezone(transport string, tzOffsetHours, max int) ([]Gateway, error) {
	gws := make([]Gateway, 0)
	gwVector := []gatewayDistance{}

	for _, gw := range p.available {
		if gw.Transport != transport {
			continue
		}
		distance := 13
		gwOffset, err := strconv.Atoi(p.locations[gw.Location].Timezone)
		if err != nil {
			log.Printf("Error sorting gateways: %v", err)
			return gws, err
		} else {
			distance = tzDistance(tzOffsetHours, gwOffset)
		}
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

func newGatewayPool(eip *eipService) *gatewayPool {
	p := gatewayPool{}
	p.available = eip.getGateways()
	p.locations = eip.Locations
	p.genLabels()
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
