package menshen

import (
	"testing"
	"time"

	"0xacab.org/leap/menshen/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLocationQualityMap(t *testing.T) {
	fast := &models.ModelsGateway{Host: "fast", IPAddr: "10.0.0.1"}
	mid := &models.ModelsGateway{Host: "mid", IPAddr: "10.0.0.2"}
	slow := &models.ModelsGateway{Host: "slow", IPAddr: "10.0.0.3"}
	dead := &models.ModelsGateway{Host: "dead", IPAddr: "10.0.0.4"}

	gwsByLocation := map[string][]*models.ModelsGateway{
		"Amsterdam": {fast, mid},
		"Seattle":   {slow},
		"Nowhere":   {dead},
	}
	rttByGw := map[*models.ModelsGateway]time.Duration{
		fast: 20 * time.Millisecond,
		mid:  40 * time.Millisecond,
		slow: 100 * time.Millisecond,
		dead: pingTimeout,
	}

	quality := buildLocationQualityMap(gwsByLocation, rttByGw)

	require.Len(t, quality, 3)
	for location, q := range quality {
		assert.GreaterOrEqual(t, q, 0.0, "quality of %s should be >= 0.0", location)
		assert.LessOrEqual(t, q, 1.0, "quality of %s should be <= 1.0", location)
	}
	assert.Equal(t, 1.0, quality["Amsterdam"], "fastest location should get quality 1.0")
	assert.Equal(t, 0.0, quality["Nowhere"], "a location with only ping timeouts should rank worst")
	assert.Less(t, quality["Seattle"], quality["Amsterdam"], "slower location should have lower quality")
}

func TestBuildLocationQualityMapSingleLocation(t *testing.T) {
	gw := &models.ModelsGateway{Host: "only", IPAddr: "10.0.0.1"}

	quality := buildLocationQualityMap(
		map[string][]*models.ModelsGateway{"Paris": {gw}},
		map[*models.ModelsGateway]time.Duration{gw: 50 * time.Millisecond},
	)

	assert.Equal(t, 1.0, quality["Paris"], "single location should get quality 1.0 (min == max)")
}

func TestBuildLocationQualityMapEmptyLocation(t *testing.T) {
	quality := buildLocationQualityMap(
		map[string][]*models.ModelsGateway{"Paris": {}},
		map[*models.ModelsGateway]time.Duration{},
	)

	assert.NotContains(t, quality, "Paris", "empty location should be skipped")
}

func TestGetBestLocationDeterministicOnTies(t *testing.T) {
	m := &Menshen{
		Gateways: []*models.ModelsGateway{
			{Host: "nuremberg-gw", IPAddr: "10.0.0.1"},
			{Host: "ashburn-gw", IPAddr: "10.0.0.2"},
		},
		locationQualityMap: map[string]float64{"nuremberg": 1.0, "ashburn": 1.0},
	}

	first, err := m.GetBestLocation("openvpn")
	require.NoError(t, err)
	assert.NotEmpty(t, first, "a tied quality map should still return a best location")

	for range 100 {
		next, err := m.GetBestLocation("openvpn")
		require.NoError(t, err)
		assert.Equal(t, first, next, "best location must not flip between calls")
	}
}
