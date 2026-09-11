package menshen

import (
	"context"
	"errors"
	"math"
	"time"

	"0xacab.org/leap/menshen/models"
	ping "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Returns true if the user selected a preferred location to connect with
func (m *Menshen) IsManualLocation() bool {
	log.Trace().Msg("Checking if a manual location is used")
	if len(m.Gateways) == 0 {
		log.Warn().Msg("The list of gateways is empty. Using auto-selection for location")
		return false
	}
	return m.userChoice != ""
}

// Returns the best location by iterating over m.locationQualityMap and finding the
// location with the highest quality
func (m *Menshen) GetBestLocation(transport string) (string, error) {
	log.Trace().
		Str("transport", transport).
		Msg("Getting best location")

	if len(m.Gateways) == 0 {
		return "", errors.New("Could not get best gateway location. The list of gateways is empty")
	}

	var bestLocation string
	bestLocationQuality := 0.0

	for location, quality := range m.locationQualityMap {
		if quality > bestLocationQuality {
			bestLocation = location
			bestLocationQuality = quality
		}
	}
	log.Debug().
		Str("location", bestLocation).
		Msg("Found best location")
	return bestLocation, nil
}

const (
	maxConcurrentPings = 16              // bounds the number of simultaneous ICMP pings
	pingTimeout        = 3 * time.Second // per-gateway ping budget; also the penalty for unreachable gateways
)

// TODO: remove function if we have a metric from menshen
func calcLatency(ip string) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return nil, err
	}

	pinger.Interval = time.Millisecond * 100
	pinger.Count = 3
	pinger.Timeout = pingTimeout
	err = pinger.Run()
	if err != nil {
		return nil, err
	}
	return pinger.Statistics(), nil
}

// measureGatewayLatencies pings all gateways concurrently (bounded by
// maxConcurrentPings) and returns one average rtt per gateway, in the same
// order as the input. Each goroutine writes to a distinct index, so no locking
// is needed. A gateway that cannot be pinged contributes pingTimeout, so
// unreachable gateways rank worst in the quality map.
func measureGatewayLatencies(gateways []*models.ModelsGateway) []time.Duration {
	rtts := make([]time.Duration, len(gateways))
	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(maxConcurrentPings)
	for i, gw := range gateways {
		g.Go(func() error {
			stats, err := calcLatency(gw.IPAddr)
			if err != nil {
				log.Warn().
					Err(err).
					Str("gateway", gw.Host).
					Msg("Could not calculate latency")
				rtts[i] = pingTimeout
				return nil
			}
			log.Trace().
				Str("gateway", gw.Host).
				Int64("rtt ms", stats.AvgRtt.Milliseconds()).
				Msg("Calculated rtt for gateway")
			rtts[i] = stats.AvgRtt
			return nil
		})
	}
	_ = g.Wait()
	return rtts
}

func (m *Menshen) GetLocationQualityMap(transport string) map[string]float64 {
	log.Trace().Msg("Getting location quality map")
	return m.locationQualityMap
}

// Updates the m.locationQualityMap struct. The struct holds the quality for each location.
// e.g. m.locationQualityMap["Paris"] = 0.3 (the higher the better), used by the GUI for visualization
// For each location we have one or more gateways. The quality values need to be floats between 0 and 1
// As we currently don't have a load metric from menshen, we just use the avg rtt of each gateway
// As GetLocationQualityMap gets called quiet often via toJson (defined in pkg/backend/status.go), we
// calculate the rtt by calling m.updateLocationQualityMap after fetching the gateways from menshen.
// GetLocationQualityMap just returns the internal value
// TODO: The rtt calculation needs to be be optimized:
//   - the code should be placed into bitmask-core (there is similar functionality, but
//     it just gives us the best host for a list of hosts based on rtt)
//
// TODO: do we need transport here as parameter?
func (m *Menshen) updateLocationQualityMap(transport string) {
	log.Debug().Msg("Calculating quality for each location")

	var all []*models.ModelsGateway
	for _, gateways := range m.gwsByLocation {
		all = append(all, gateways...)
	}
	rtts := measureGatewayLatencies(all)

	rttByGw := make(map[*models.ModelsGateway]time.Duration, len(all))
	for i, gw := range all {
		rttByGw[gw] = rtts[i]
	}

	m.locationQualityMap = buildLocationQualityMap(m.gwsByLocation, rttByGw)
}

// buildLocationQualityMap averages the per-gateway rtts per location and
// normalizes the values to [0, 1] (higher is better).
/*
	implementation description:
		1) iterate over gwsByLocation => gives us location and a list of gateways
		2) for each location, calculate the average rtt for all gateways

	normalization:
		- if we have rtt values, we need to normalize them (get floats between 0 and 1)
		- formulae used: https://www.statology.org/normalize-data-between-0-and-1/
		- therefore, we need to find out the min and max of all avgRtts for all locations
		- the algorithm has drawbacks: the worst location always gets a value of 0, the best
		  a vlaue of 1 - independent of the actual rtt (can be very high)
		- TODO: check algorithm of v3 implementation
*/
func buildLocationQualityMap(gwsByLocation map[string][]*models.ModelsGateway, rttByGw map[*models.ModelsGateway]time.Duration) map[string]float64 {
	qualityMap := make(map[string]float64)
	minAvgRtt := math.MaxFloat64
	maxAvgRtt := 0.0

	for location, gateways := range gwsByLocation {
		if len(gateways) == 0 {
			continue
		}
		sumLocation := 0.0
		for _, gw := range gateways {
			sumLocation += float64(rttByGw[gw].Milliseconds())
		}
		locationRttAvg := sumLocation / float64(len(gateways))
		qualityMap[location] = locationRttAvg

		if locationRttAvg < minAvgRtt {
			minAvgRtt = locationRttAvg
		}
		if locationRttAvg > maxAvgRtt {
			maxAvgRtt = locationRttAvg
		}
	}

	log.Trace().
		Msgf("location quality map: %v", qualityMap)

	// normalize values (from rtt in ms to a number between 0 and 1)
	for location, rtt := range qualityMap {
		avgRttNormalized := (rtt - minAvgRtt) / (maxAvgRtt - minAvgRtt)
		if math.IsNaN(avgRttNormalized) {
			avgRttNormalized = 0
		}
		// higher latency is bad, so 1 - avgRttNormalized
		qualityMap[location] = 1 - avgRttNormalized
	}
	log.Trace().
		Msgf("location quality map normalized: %v", qualityMap)

	return qualityMap
}

// Returns a map[string][string] with gateway locations and their country code.
// Only used for the GUI
// locationLabels["Paris"] = ["Paris", "FR"]
// locationLabels["Seattle"] = ["Seattle", "US"]
// This functions gets called quiet often via toJson function in pkg/backend/status.go
// TODO: use a smarter structure if we get rid of v3 (needs to be cpp compatible)
func (m *Menshen) GetLocationLabels(transport string) map[string][]string {
	log.Trace().Msg("Building location label map")
	locationLabels := make(map[string][]string)

	service, err := m.GetService()
	if err != nil {
		return locationLabels
	}

	for _, gw := range m.Gateways {
		_, exist := locationLabels[gw.Location]
		if !exist {
			countryCode := getCountryCodeForLocation(gw.Location, service)
			locationDisplayName := getLocationName(gw.Location, service)
			locationLabels[gw.Location] = []string{locationDisplayName, countryCode}
		}
	}
	return locationLabels
}

func getLocationName(location string, service *models.ModelsEIPService) string {
	l, set := service.Locations[location]
	if !set {
		return "Unknown"
	}

	return l.DisplayName
}

func getCountryCodeForLocation(location string, service *models.ModelsEIPService) string {
	l, set := service.Locations[location]
	if !set {
		return ""
	}
	return l.CountryCode
}
