package backend

import (
	"bytes"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	bitmaskAutostart "0xacab.org/leap/bitmask-vpn/pkg/bitmask/autostart"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"0xacab.org/leap/bitmask-vpn/pkg/snowflake"
)

const (
	offStr      = "off"
	startingStr = "starting"
	onStr       = "on"
	stoppingStr = "stopping"
	failedStr   = "failed"
	unknownStr  = "unknown"
)

// ctx will be our glorious global object.
// if we ever switch again to a provider-agnostic app, we should keep a map here.
var ctx *connectionCtx

// these mutexes protect setting and updating the global status in this go backend
var statusMutex sync.Mutex
var updateMutex sync.Mutex

// The connectionCtx keeps the global state that is passed around to C-land. It
// also serves as the primary way of passing requests from the frontend to the
// Go-core, by letting the UI write some of these variables and processing
// them.

type connectionCtx struct {
	AppName           string              `json:"appName"`
	Provider          string              `json:"provider"`
	TosURL            string              `json:"tosURL"`
	HelpURL           string              `json:"helpURL"`
	AskForDonations   bool                `json:"askForDonations"`
	DonateDialog      bool                `json:"donateDialog"`
	DonateURL         string              `json:"donateURL"`
	LoginDialog       bool                `json:"loginDialog"`
	LoginOk           bool                `json:"loginOk"`
	Version           string              `json:"version"`
	Errors            string              `json:"errors"`
	Status            status              `json:"status"`
	Locations         map[string]float64  `json:"locations"`
	Providers         []string            `json:"providers"`
	LocationLabels    map[string][]string `json:"locationLabels"`
	CurrentGateway    string              `json:"currentGateway"`
	CurrentLocation   string              `json:"currentLocation"`
	CurrentCountry    string              `json:"currentCountry"`
	BestLocation      string              `json:"bestLocation"`
	Transport         string              `json:"transport"`
	UseUDP            bool                `json:"udp"`
	OffersUDP         bool                `json:"offersUdp"`
	ManualLocation    bool                `json:"manualLocation"`
	IsReady           bool                `json:"isReady"`
	CanUpgrade        bool                `json:"canUpgrade"`
	Motd              string              `json:"motd"`
	HasTor            bool                `json:"hasTor"`
	UseSnowflake      bool                `json:"snowflake"`
	SnowflakeProgress int                 `json:"snowflakeProgress"`
	SnowflakeTag      string              `json:"snowflakeTag"`
	OffersObfs4       bool                `json:"offersObfs4"`
	OffersQUIC        bool                `json:"offersQuic"`
	OffersKCP         bool                `json:"offersKcp"`
	OffersHopping     bool                `json:"offersHopping"`
	bm                bitmask.Bitmask
	autostart         bitmaskAutostart.Autostart
	cfg               *config.Config
}

func (c *connectionCtx) toJson() ([]byte, error) {
	statusMutex.Lock()
	if c.bm != nil {
		transport := c.bm.GetTransport()
		c.Locations = c.bm.GetLocationQualityMap(transport)
		c.LocationLabels = c.bm.GetLocationLabels(transport)
		c.CurrentGateway = c.bm.GetCurrentGateway()
		c.CurrentLocation = c.bm.GetCurrentLocation()
		c.CurrentCountry = c.bm.GetCurrentCountry()
		c.BestLocation = c.bm.GetBestLocation(transport)
		c.Transport = transport
		c.UseUDP = c.cfg.UDP // TODO initialize bitmask param?
		c.OffersUDP = c.bm.OffersUDP()
		c.UseSnowflake = c.cfg.Snowflake // TODO initialize bitmask
		c.ManualLocation = c.bm.IsManualLocation()
		c.CanUpgrade = c.bm.CanUpgrade()
		c.Motd = c.bm.GetMotd()
		c.HasTor = snowflake.HasTor()
		c.OffersObfs4 = c.bm.OffersObfs4()
		c.OffersKCP = c.bm.OffersKCP()
		c.OffersQUIC = c.bm.OffersQUIC()
		c.OffersHopping = c.bm.OffersHopping()
	}
	defer statusMutex.Unlock()
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c connectionCtx) updateStatus() {
	updateMutex.Lock()
	defer func() {
		log.Debug().
			Msg("Exiting from the updateStatus go routine checking the values of the statusCh")
		updateMutex.Unlock()
	}()
	if stStr, err := c.bm.GetStatus(); err != nil {
		log.Error().
			Err(err).
			Msg("Could not get OpenVPN status")
		setStatusFromStr(offStr)
	} else {
		setStatusFromStr(stStr)
	}

	go func() {
		snowflakeCh := c.bm.GetSnowflakeCh()
		for {
			select {
			case event := <-snowflakeCh:
				setSnowflakeStatus(event)
			}
		}
	}()

	statusCh := c.bm.GetStatusCh()
	statusCloseCh := c.bm.GetStatusCloseCh()
	log.Debug().
		Msg("Going to enter the statusCh value check loop")
	for {
		select {
		case stStr := <-statusCh:
			log.Debug().
				Str("statusCh", stStr).
				Msg("Got value in the statusCh")
			if stStr == "" {
				setStatusFromStr(offStr)
				continue
			}
			setStatusFromStr(stStr)
		case _ = <-statusCloseCh:
			log.Debug().
				Msg("Closing the statusCh reading loop via the statusCloseCh")
			setStatusFromStr(offStr)
			return
		}
	}
}

func setSnowflakeStatus(event *snowflake.StatusEvent) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	ctx.SnowflakeProgress = event.Progress
	ctx.SnowflakeTag = event.Tag
	go trigger(OnStatusChanged)
}

func setStatus(st status) {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	ctx.Status = st
	go trigger(OnStatusChanged)
}

// the status type reflects the current VPN status. Go code is responsible for updating
// it; the C gui just watches its changes and pulls its updates via the serialized
// context object.

type status int

const (
	off status = iota
	starting
	on
	stopping
	failed
	unknown
)

func (s status) String() string {
	return [...]string{offStr, startingStr, onStr, stoppingStr, failedStr, unknownStr}[s]
}

func (s status) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(s.String())
	b.WriteString(`"`)
	return b.Bytes(), nil
}

func fromString(st string) status {
	switch st {
	case offStr:
		return off
	case startingStr:
		return starting
	case onStr:
		return on
	case stoppingStr:
		return stopping
	case failedStr:
		return failed
	default:
		return unknown
	}
}

func setStatusFromStr(stStr string) {
	log.Trace().
		Str("status", stStr).
		Msg("Setting GUI status")
	setStatus(fromString(stStr))
}
