package menshen

import (
	"errors"
	"math"
	"strings"
	"time"

	ping "github.com/prometheus-community/pro-bing"
	"github.com/rs/zerolog/log"
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

// TODO: remove function if we have a metric from menshen
func calcLatency(ip string) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		return nil, err
	}

	pinger.Interval = time.Millisecond * 100
	pinger.Count = 3
	pinger.Timeout = time.Second * 3
	err = pinger.Run()
	if err != nil {
		return nil, err
	}
	return pinger.Statistics(), nil
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
//   - calculating the rtt should be parallized
//
// TODO: do we need transport here as parameter?
func (m *Menshen) updateLocationQualityMap(transport string) {
	log.Debug().Msg("Calculating quality for each location")
	/*
		implementation description:
			1) iterate over m.gwsByLocation => gives us location and a list of gateways
			2) for each gateway, calculate the rtt
			3) for each location, calculate the average rtt for all gateways

		normalization:
			- if we have rtt values, we need to normalize them (get floats between 0 and 1)
			- formulae used: https://www.statology.org/normalize-data-between-0-and-1/
			- therefore, we need to find out the min and max of all avgRtts for all locations
			- the algorithm has drawbacks: the worst location always gets a value of 0, the best
			  a vlaue of 1 - independent of the actual rtt (can be very high)
			- TODO: check algorithm of v3 implementation
	*/
	qualityMap := make(map[string]float64)
	minAvgRtt := math.MaxFloat64
	maxAvgRtt := 0.0

	for location, gateways := range m.gwsByLocation {
		sum_location := int64(0)
		counter_location := int64(0)

		for _, gw := range gateways {
			stats, err := calcLatency(gw.IPAddr)
			if err != nil {
				log.Warn().
					Err(err).
					Str("gateway", gw.Host).
					Msg("Could not calculate latency")
				sum_location += math.MaxInt64
			} else {
				log.Trace().
					Str("location", location).
					Str("gateway", gw.Host).
					Int64("rtt ms", stats.AvgRtt.Milliseconds()).
					Msg("Calculated rtt for gateway")
				sum_location += stats.AvgRtt.Milliseconds()
			}
			counter_location += 1
		}

		locationRttAvg := float64(sum_location / counter_location)
		qualityMap[strings.Title(location)] = locationRttAvg

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
		// higher latency is bad, so 1 - avgRttNormalized
		qualityMap[location] = 1 - avgRttNormalized
	}
	log.Trace().
		Msgf("location quality map normalized: %v", qualityMap)

	m.locationQualityMap = qualityMap
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

	for _, gw := range m.Gateways {
		_, exist := locationLabels[gw.Host]
		if !exist {
			countryCode := getCountryCodeForLocation(gw.Location)
			// TODO: get rid of strings.Title if menshen supports gateway identifier
			locationLabels[strings.Title(gw.Location)] = []string{strings.Title(gw.Location), countryCode}
		}
	}
	return locationLabels
}

// Returns the CountryCode for a gateway location
// TODO: remove this if menshen has support for CountryCode
func getCountryCodeForLocation(location string) string {
	switch location {
	case "paris":
		return "FR"
	case "seattle":
		return "US"
	case "miami":
		return "US"
	case "newyorkcity":
		return "US"
	case "montreal":
		return "US"
	case "amsterdam":
		return "NL"
	}
	return "TODO: CC"

}
