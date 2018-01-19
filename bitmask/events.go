package bitmask

import (
	"github.com/pebbe/zmq4"
)

const (
	eventsEndpoint = "tcp://127.0.0.1:9001"
	//serverKeyPath  = "/home/user/.config/leap/events/zmq_certificates/public_keys/server.key" // FIXME
)

func initEvents() (*zmq4.Socket, error) {
	socket, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}

	if zmq4.HasCurve() {
		// TODO
	}

	err = socket.Connect(eventsEndpoint)
	if err != nil {
		return nil, err
	}
	return socket, nil
}

func (b *Bitmask) fetchStatus() {
	// TODO: this should be a subscription to the event
	for {
		time.Sleep(time.Second)
		status, err := b.GetStatus()
		if err != nil {
			log.Printf("Error receiving status: %v", err)
			continue
		}
		b.statusCh <- status
	}
}
