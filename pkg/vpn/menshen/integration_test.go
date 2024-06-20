package menshen

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
}

func getMenshenInstance(t *testing.T) *Menshen {
	// TODO: Skip tests during CI https://0xacab.org/leap/bitmask-vpn/-/issues/826
	if os.Getenv("CI") != "" {
		t.Skip("Not running integration tests right now in the CI")
	}
	m, err := New()
	require.NoError(t, err, "Could not create menshen instance")
	return m
}

func TestGetAllGateways(t *testing.T) {
	// needs API_URL="http://localhost:8443" via env
	m := getMenshenInstance(t)
	gateways, err := m.GetAllGateways("openvpn")
	require.NoError(t, err, "GetAllGateways returned an error")
	assert.Greater(t, len(gateways), 0, "There should multiple gateways fetched")
	log.Info().
		Int("gateways", len(gateways)).
		Msg("Got gateways")
}

func TestGetCert(t *testing.T) {
	m := getMenshenInstance(t)
	certBytes, err := m.GetPemCertificate()
	cert := string(certBytes)
	require.NoError(t, err, "GetPemCertificates returned an error")
	assert.Contains(t, cert, " PRIVATE KEY-----")
	assert.Contains(t, cert, "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, cert, "-----END CERTIFICATE-----")
	log.Info().Msgf("Got valid client certificate: \n%v", cert)
}

func TestGetVpnArguments(t *testing.T) {
	m := getMenshenInstance(t)
	args, err := m.GetOpenvpnArgs()
	require.NoError(t, err, "GetOpenvpnArgs returned an error")
	assert.Contains(t, args, "--dev")
	assert.Contains(t, args, "tun")
	assert.Contains(t, args, "--persist-key") // comes as bool from api in json
	log.Info().Msgf("Got valid OpenVPN arguments: %v", args)

}

func TestLatency(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Not running integration tests right now in the CI")
	}
	ip := "1.1.1.1"
	stats, err := calcLatency(ip)
	require.NoError(t, err, "Could not calc latency")
	log.Info().
		Str("ip", ip).
		Int64("AvgRtt", stats.AvgRtt.Milliseconds()).
		Msg("Calculated latency")
}

func TestLocationQualityMap(t *testing.T) {
	m := getMenshenInstance(t)

	_, err := m.GetAllGateways("openvpn")
	require.NoError(t, err, "GetAllGateways returned an error")

	locationQualtyMap := m.GetLocationQualityMap("openvpn")
	for _, quality := range locationQualtyMap {
		assert.GreaterOrEqual(t, quality, 0.0, "quality should be higher than 0.0")
		assert.LessOrEqual(t, quality, 1.0, "quality should be lower than 1.0")
	}
}

func TestLocationLabels(t *testing.T) {
	m := getMenshenInstance(t)
	_, err := m.GetAllGateways("openvpn")
	require.NoError(t, err, "GetAllGateways returned an error")

	labelMap := m.GetLocationLabels("transport")
	for _, city := range labelMap {
		log.Info().
			Str("location", city[0]).
			Str("country", city[1]).
			Msg("Got location")
	}
}

// Get all locations and qualities with GetLocationQualityMap
// Get best location by caling GetBestLocation
// check that the quality is really the best
func TestGetBestLocation(t *testing.T) {
	m := getMenshenInstance(t)

	_, err := m.GetAllGateways("openvpn")
	require.NoError(t, err, "GetAllGateways returned an error")

	locationQualtyMap := m.GetLocationQualityMap("openvpn")

	location, err := m.GetBestLocation("openvpn")
	require.NoError(t, err, "GetBestLocation returned an error")

	bestQuality, exist := locationQualtyMap[location]
	require.True(t, exist, "location was not found in qualityMap")

	for _, quality := range locationQualtyMap {
		assert.GreaterOrEqual(t, bestQuality, quality, "bestQuality should be higher or equal")
	}

}

// Test if gateways are shuffled - GetBestGateways() should not return the
// same gateways twice
func TestGetBestGatewaysShuffled(t *testing.T) {
	transport := "openvpn"
	m := getMenshenInstance(t)

	_, err := m.GetAllGateways(transport)
	assert.NoError(t, err, "GetAllGateways returned an error")

	location, err := m.GetBestLocation(transport)
	assert.NoError(t, err, "m.GetBestLocation returned an error")

	m.SetManualGateway(location)

	gws1, err := m.GetBestGateways(transport)
	log.Info().Msgf("gateways first: %v", gws1)
	require.NoError(t, err, "GetBestGateways returned an error")

	gws2, err := m.GetBestGateways(transport)
	log.Info().Msgf("gateways second: %v", gws2)
	require.NoError(t, err, "GetBestGateways returned an error")

	// this does not work: maybe we only have one gateway per location
	// sorting one element does not work
	// we could loop and ask menshen for gateways until we have >1 gateways
	// or just skip if we have just one per location
	if len(gws1) < 2 || len(gws2) < 2 {
		log.Warn().Msg("Can not test shuffled gateways. There are not enough gateways returned by menshen to shuffle/compare")
		return
	}
	gws1Hosts := []string{gws1[0].Host, gws1[1].Host}
	gws2Hosts := []string{gws2[0].Host, gws2[1].Host}

	log.Info().
		Str("gw1Hosts", strings.Join(gws1Hosts, " ")).
		Str("gw2Hosts", strings.Join(gws2Hosts, " ")).
		Msg("Asked menshen twice for gateways. Checking order...")

	if reflect.DeepEqual(gws1, gws2) {
		log.Warn().Msg("Soft Fail: Gateways should be shuffled and not in the same order. This can happen (if we have only two gateways, shuffeling them can result in the same order")
	}

}
