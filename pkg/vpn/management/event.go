// Copyright (c) 2016 Martin Atkins
// Copyright (c) 2021 LEAP Encryption Access Project

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package management

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var eventSep = []byte(":")
var fieldSep = []byte(",")
var byteCountEventKW = []byte("BYTECOUNT")
var byteCountCliEventKW = []byte("BYTECOUNT_CLI")
var clientEventKW = []byte("CLIENT")
var echoEventKW = []byte("ECHO")
var fatalEventKW = []byte("FATAL")
var holdEventKW = []byte("HOLD")
var infoEventKW = []byte("INFO")
var logEventKW = []byte("LOG")
var needOkEventKW = []byte("NEED-OK")
var needStrEventKW = []byte("NEED-STR")
var passwordEventKW = []byte("PASSWORD")
var stateEventKW = []byte("STATE")

type Event interface {
	String() string
}

// UnknownEvent represents an event of a type that this package doesn't
// know about.
//
// Future versions of this library may learn about new event types, so a
// caller should exercise caution when making use of events of this type
// to access unsupported behavior. Backward-compatibility is *not*
// guaranteed for events of this type.
type UnknownEvent struct {
	keyword []byte
	body    []byte
}

func (e *UnknownEvent) Type() string {
	return string(e.keyword)
}

func (e *UnknownEvent) Body() string {
	return string(e.body)
}

func (e *UnknownEvent) String() string {
	return fmt.Sprintf("%s: %s", e.keyword, e.body)
}

// MalformedEvent represents a message from the OpenVPN process that is
// presented as an event but does not comply with the expected event syntax.
//
// Events of this type should never be seen but robust callers will accept
// and ignore them, possibly generating some kind of debugging message.
//
// One reason for potentially seeing events of this type is when the target
// program is actually not an OpenVPN process at all, but in fact this client
// has been connected to a different sort of server by mistake.
type MalformedEvent struct {
	raw []byte
}

func (e *MalformedEvent) String() string {
	return fmt.Sprintf("Malformed Event %q", e.raw)
}

// HoldEvent is a notification that the OpenVPN process is in a management
// hold and will not continue connecting until the hold is released, e.g.
// by calling client.HoldRelease()
type HoldEvent struct {
	body []byte
}

func (e *HoldEvent) String() string {
	return string(e.body)
}

// StateEvent is a notification of a change of connection state. It can be
// used, for example, to detect if the OpenVPN connection has been interrupted
// and the OpenVPN process is attempting to reconnect.
type StateEvent struct {
	body []byte

	// bodyParts is populated only on first request, giving us the
	// separate comma-separated elements of the message. Not all
	// fields are populated for all states.
	bodyParts [][]byte
}

func (e *StateEvent) RawTimestamp() string {
	parts := e.parts()
	return string(parts[0])
}

func (e *StateEvent) NewState() string {
	parts := e.parts()
	return string(parts[1])
}

func (e *StateEvent) Description() string {
	parts := e.parts()
	return string(parts[2])
}

// LocalTunnelAddr returns the IP address of the local interface within
// the tunnel, as a string that can be parsed using net.ParseIP.
//
// This field is only populated for events whose NewState returns
// either ASSIGN_IP or CONNECTED.
func (e *StateEvent) LocalTunnelAddr() string {
	parts := e.parts()
	return string(parts[3])
}

// RemoteAddr returns the non-tunnel IP address of the remote
// system that has connected to the local OpenVPN process.
//
// This field is only populated for events whose NewState returns
// CONNECTED.
func (e *StateEvent) RemoteAddr() string {
	parts := e.parts()
	return string(parts[4])
}

// RemotePort returns the port of the remote openvpn process.
// This field is only populated for events whose NewState returns
// CONNECTED.
func (e *StateEvent) RemotePort() string {
	parts := e.parts()
	// parts[5] is "80,,,fd15:53b6:dead::2", 80 is the port
	return strings.Split(string(parts[5]), ",")[0]
}

func (e *StateEvent) String() string {
	newState := e.NewState()
	switch newState {
	case "ASSIGN_IP":
		return fmt.Sprintf("%s: %s", newState, e.LocalTunnelAddr())
	case "CONNECTED":
		return fmt.Sprintf("%s: %s:%s", newState, e.RemoteAddr(), e.RemotePort())
	default:
		desc := e.Description()
		if desc != "" {
			return fmt.Sprintf("%s: %s", newState, desc)
		} else {
			return newState
		}
	}
}

func (e *StateEvent) parts() [][]byte {
	if e.bodyParts == nil {
		// State messages currently have only five segments, but
		// we'll ask for 5 so any additional fields that might show
		// up in newer versions will gather in an element we're
		// not actually using.
		e.bodyParts = bytes.SplitN(e.body, fieldSep, 6)

		// Prevent crash if the server has sent us a malformed
		// status message. This should never actually happen if
		// the server is behaving itself.
		if len(e.bodyParts) < 6 {
			expanded := make([][]byte, 6)
			copy(expanded, e.bodyParts)
			e.bodyParts = expanded
		}
	}
	return e.bodyParts
}

// EchoEvent is emitted by an OpenVPN process running in client mode when
// an "echo" command is pushed to it by the server it has connected to.
//
// The format of the echo message is free-form, since this message type is
// intended to pass application-specific data from the server-side config
// into whatever client is consuming the management prototcol.
//
// This event is emitted only if the management client has turned on events
// of this type using client.SetEchoEvents(true)
type EchoEvent struct {
	body []byte
}

func (e *EchoEvent) RawTimestamp() string {
	sepIndex := bytes.Index(e.body, fieldSep)
	if sepIndex == -1 {
		return ""
	}
	return string(e.body[:sepIndex])
}

func (e *EchoEvent) Message() string {
	sepIndex := bytes.Index(e.body, fieldSep)
	if sepIndex == -1 {
		return ""
	}
	return string(e.body[sepIndex+1:])
}

func (e *EchoEvent) String() string {
	return fmt.Sprintf("ECHO: %s", e.Message())
}

// ByteCountEvent represents a periodic snapshot of data transfer in bytes
// on a VPN connection.
//
// For OpenVPN *servers*, events are emitted for each client and the method
// ClientId identifies thet client.
//
// For other OpenVPN modes, events are emitted only once per interval for the
// single connection managed by the target process, and ClientId returns
// the empty string.
type ByteCountEvent struct {
	hasClient bool
	body      []byte

	// populated on first call to parts()
	bodyParts [][]byte
}

func (e *ByteCountEvent) ClientId() string {
	if !e.hasClient {
		return ""
	}

	return string(e.parts()[0])
}

func (e *ByteCountEvent) BytesIn() int {
	index := 0
	if e.hasClient {
		index = 1
	}
	str := string(e.parts()[index])
	val, _ := strconv.Atoi(str)
	// Ignore error, since this should never happen if OpenVPN is
	// behaving itself.
	return val
}

func (e *ByteCountEvent) BytesOut() int {
	index := 1
	if e.hasClient {
		index = 2
	}
	str := string(e.parts()[index])
	val, _ := strconv.Atoi(str)
	// Ignore error, since this should never happen if OpenVPN is
	// behaving itself.
	return val
}

func (e *ByteCountEvent) String() string {
	if e.hasClient {
		return fmt.Sprintf("Client %s: %d in, %d out", e.ClientId(), e.BytesIn(), e.BytesOut())
	} else {
		return fmt.Sprintf("%d in, %d out", e.BytesIn(), e.BytesOut())
	}
}

func (e *ByteCountEvent) parts() [][]byte {
	if e.bodyParts == nil {
		e.bodyParts = bytes.SplitN(e.body, fieldSep, 4)

		wantCount := 2
		if e.hasClient {
			wantCount = 3
		}

		// Prevent crash if the server has sent us a malformed
		// message. This should never actually happen if the
		// server is behaving itself.
		if len(e.bodyParts) < wantCount {
			expanded := make([][]byte, wantCount)
			copy(expanded, e.bodyParts)
			e.bodyParts = expanded
		}
	}
	return e.bodyParts
}

func upgradeEvent(raw []byte) Event {
	splitIdx := bytes.Index(raw, eventSep)
	if splitIdx == -1 {
		// Should never happen, but we'll handle it robustly if it does.
		return &MalformedEvent{raw}
	}

	keyword := raw[:splitIdx]
	body := raw[splitIdx+1:]

	switch {
	case bytes.Equal(keyword, stateEventKW):
		return &StateEvent{body: body}
	case bytes.Equal(keyword, holdEventKW):
		return &HoldEvent{body}
	case bytes.Equal(keyword, echoEventKW):
		return &EchoEvent{body}
	case bytes.Equal(keyword, byteCountEventKW):
		return &ByteCountEvent{hasClient: false, body: body}
	case bytes.Equal(keyword, byteCountCliEventKW):
		return &ByteCountEvent{hasClient: true, body: body}
	default:
		return &UnknownEvent{keyword, body}
	}
}
