package backend

import (
	"bytes"
	"encoding/json"
	"log"

	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

const (
	offStr      = "off"
	startingStr = "starting"
	onStr       = "on"
	stoppingStr = "stopping"
	failedStr   = "failed"
)

// ctx will be our glorious global object.
// if we ever switch again to a provider-agnostic app, we should keep a map here.
var ctx *connectionCtx

// The connectionCtx keeps the global state that is passed around to C-land. It
// also serves as the primary way of passing requests from the frontend to the
// Go-core, by letting the UI write some of these variables and processing
// them.

type connectionCtx struct {
	AppName         string `json:"appName"`
	Provider        string `json:"provider"`
	TosURL          string `json:"tosURL"`
	HelpURL         string `json:"helpURL"`
	AskForDonations bool   `json:"askForDonations"`
	DonateDialog    bool   `json:"donateDialog"`
	DonateURL       string `json:"donateURL"`
	Version         string `json:"version"`
	Status          status `json:"status"`
	bm              bitmask.Bitmask
	cfg             *config.Config
}

func (c connectionCtx) toJson() ([]byte, error) {
	stmut.Lock()
	defer stmut.Unlock()
	b, err := json.Marshal(c)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return b, nil
}

func (c connectionCtx) updateStatus() {
	if stStr, err := c.bm.GetStatus(); err != nil {
		log.Printf("Error getting status: %v", err)
	} else {
		setStatusFromStr(stStr)
	}

	statusCh := c.bm.GetStatusCh()
	for {
		select {
		case stStr := <-statusCh:
			setStatusFromStr(stStr)
		}
	}
}

func setStatus(st status) {
	stmut.Lock()
	defer stmut.Unlock()
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
	return [...]string{offStr, startingStr, onStr, stoppingStr, failedStr}[s]
}

func (s status) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(s.String())
	b.WriteString(`"`)
	return b.Bytes(), nil
}

func (s status) fromString(st string) status {
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
	setStatus(unknown.fromString(stStr))
}
