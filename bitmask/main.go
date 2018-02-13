package bitmask

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/pebbe/zmq4"
)

const (
	// On win should be: tcp://127.0.0.1:5001
	coreEndpoint = "ipc:///tmp/bitmask.core.sock"
)

// Bitmask holds the bitmask client data
type Bitmask struct {
	coresoc  *zmq4.Socket
	eventsoc *zmq4.Socket
	statusCh chan string
}

// Init the connection to bitmask
func Init() (*Bitmask, error) {
	statusCh := make(chan string)
	coresoc, err := initCore()
	if err != nil {
		return nil, err
	}
	eventsoc, err := initEvents()
	if err != nil {
		return nil, err
	}

	b := Bitmask{coresoc, eventsoc, statusCh}
	go b.eventsHandler()
	return &b, nil
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() chan string {
	return b.statusCh
}

// Close the connection to bitmask
func (b *Bitmask) Close() {
	_, err := b.send("core", "stop")
	if err != nil {
		log.Printf("Got an error stopping bitmaskd: %v", err)
	}
	b.coresoc.Close()
}

func (b *Bitmask) send(parts ...interface{}) (map[string]interface{}, error) {
	_, err := b.coresoc.SendMessage(parts...)
	if err != nil {
		return nil, err
	}
	resJSON, err := b.coresoc.RecvBytes(0)
	if err != nil {
		return nil, err
	}
	return parseResponse(resJSON)
}

func parseResponse(resJSON []byte) (map[string]interface{}, error) {
	var response struct {
		Result map[string]interface{}
		Error  string
	}
	err := json.Unmarshal(resJSON, &response)
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return response.Result, err
}

func initCore() (*zmq4.Socket, error) {
	socket, err := zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		return nil, err
	}

	err = socket.Connect(coreEndpoint)
	return socket, err
}
