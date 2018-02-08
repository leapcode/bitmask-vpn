package bitmask

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/pebbe/zmq4"
)

const (
	eventsEndpoint = "tcp://127.0.0.1:9001"
	statusEvent    = "VPN_STATUS_CHANGED"
)

func initEvents() (*zmq4.Socket, error) {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	if zmq4.HasCurve() {
		err = initCurve(socket)
		if err != nil {
			return nil, err
		}
	}

	err = socket.Connect(eventsEndpoint)
	if err != nil {
		return nil, err
	}

	err = socket.SetSubscribe(statusEvent)
	return socket, err
}

func initCurve(socket *zmq4.Socket) error {
	serverKeyData, err := ioutil.ReadFile(getServerKeyPath())
	if err != nil {
		return err
	}

	pubkey, seckey, err := zmq4.NewCurveKeypair()
	if err != nil {
		return err
	}

	serverkey := strings.Split(string(serverKeyData), "\"")[1]
	return socket.ClientAuthCurve(serverkey, pubkey, seckey)
}

func (b *Bitmask) eventsHandler() {
	for {
		msg, err := b.eventsoc.RecvMessage(0)
		if err != nil {
			break
		}
		log.Println(msg[0])
		if msg[0][:len(statusEvent)] != statusEvent {
			continue
		}

		status, err := b.GetStatus()
		if err != nil {
			log.Printf("Error receiving status: %v", err)
			continue
		}
		b.statusCh <- status
	}
}

func getServerKeyPath() string {
	return filepath.Join(ConfigPath, "events", "zmq_certificates", "public_keys", "server.key")
}
