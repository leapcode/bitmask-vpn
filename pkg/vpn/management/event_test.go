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
	"fmt"
	"testing"
)

// A key requirement of our event parsing is that it must never cause a
// panic, even if the OpenVPN process sends us malformed garbage.
//
// Therefore most of the tests in here are testing various tortured error
// cases, which are all expected to produce an event object, though the
// contents of that event object will be nonsensical if the OpenVPN server
// sends something nonsensical.

func TestMalformedEvent(t *testing.T) {
	testCases := [][]byte{
		[]byte(""),
		[]byte("HTTP/1.1 200 OK"),
		[]byte("     "),
		[]byte("\x00"),
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase)

		var malformed *MalformedEvent
		var ok bool
		if malformed, ok = event.(*MalformedEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, malformed)
			continue
		}

		wantString := fmt.Sprintf("Malformed Event %q", testCase)
		if gotString := malformed.String(); gotString != wantString {
			t.Errorf("test %d String returned %q; want %q", i, gotString, wantString)
		}
	}
}

func TestUnknownEvent(t *testing.T) {
	type TestCase struct {
		Input    []byte
		WantType string
		WantBody string
	}
	testCases := []TestCase{
		{
			Input:    []byte("DUMMY:baz"),
			WantType: "DUMMY",
			WantBody: "baz",
		},
		{
			Input:    []byte("DUMMY:"),
			WantType: "DUMMY",
			WantBody: "",
		},
		{
			Input:    []byte("DUMMY:abc,123,456"),
			WantType: "DUMMY",
			WantBody: "abc,123,456",
		},
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase.Input)

		var unk *UnknownEvent
		var ok bool
		if unk, ok = event.(*UnknownEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, unk)
			continue
		}

		if got, want := unk.Type(), testCase.WantType; got != want {
			t.Errorf("test %d Type returned %q; want %q", i, got, want)
		}
		if got, want := unk.Body(), testCase.WantBody; got != want {
			t.Errorf("test %d Body returned %q; want %q", i, got, want)
		}
	}
}

func TestHoldEvent(t *testing.T) {
	testCases := [][]byte{
		[]byte("HOLD:"),
		[]byte("HOLD:waiting for hold release"),
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase)

		var hold *HoldEvent
		var ok bool
		if hold, ok = event.(*HoldEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, hold)
			continue
		}
	}
}

func TestEchoEvent(t *testing.T) {
	type TestCase struct {
		Input         []byte
		WantTimestamp string
		WantMessage   string
	}
	testCases := []TestCase{
		{
			Input:         []byte("ECHO:123,foo"),
			WantTimestamp: "123",
			WantMessage:   "foo",
		},
		{
			Input:         []byte("ECHO:123,"),
			WantTimestamp: "123",
			WantMessage:   "",
		},
		{
			Input:         []byte("ECHO:,foo"),
			WantTimestamp: "",
			WantMessage:   "foo",
		},
		{
			Input:         []byte("ECHO:,"),
			WantTimestamp: "",
			WantMessage:   "",
		},
		{
			Input:         []byte("ECHO:"),
			WantTimestamp: "",
			WantMessage:   "",
		},
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase.Input)

		var echo *EchoEvent
		var ok bool
		if echo, ok = event.(*EchoEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, echo)
			continue
		}

		if got, want := echo.RawTimestamp(), testCase.WantTimestamp; got != want {
			t.Errorf("test %d RawTimestamp returned %q; want %q", i, got, want)
		}
		if got, want := echo.Message(), testCase.WantMessage; got != want {
			t.Errorf("test %d Message returned %q; want %q", i, got, want)
		}
	}
}

func TestStateEvent(t *testing.T) {
	type TestCase struct {
		Input          []byte
		WantTimestamp  string
		WantState      string
		WantDesc       string
		WantLocalAddr  string
		WantRemoteAddr string
		WantRemotePort string
	}
	testCases := []TestCase{
		{
			Input:          []byte("STATE:"),
			WantTimestamp:  "",
			WantState:      "",
			WantDesc:       "",
			WantLocalAddr:  "",
			WantRemoteAddr: "",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:,"),
			WantTimestamp:  "",
			WantState:      "",
			WantDesc:       "",
			WantLocalAddr:  "",
			WantRemoteAddr: "",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:,,,,"),
			WantTimestamp:  "",
			WantState:      "",
			WantDesc:       "",
			WantLocalAddr:  "",
			WantRemoteAddr: "",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:123,CONNECTED,good,172.16.0.1,192.168.4.1"),
			WantTimestamp:  "123",
			WantState:      "CONNECTED",
			WantDesc:       "good",
			WantLocalAddr:  "172.16.0.1",
			WantRemoteAddr: "192.168.4.1",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:123,RECONNECTING,SIGHUP,,"),
			WantTimestamp:  "123",
			WantState:      "RECONNECTING",
			WantDesc:       "SIGHUP",
			WantLocalAddr:  "",
			WantRemoteAddr: "",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:123,RECONNECTING,SIGHUP,,,"),
			WantTimestamp:  "123",
			WantState:      "RECONNECTING",
			WantDesc:       "SIGHUP",
			WantLocalAddr:  "",
			WantRemoteAddr: "",
			WantRemotePort: "",
		},
		{
			Input:          []byte("STATE:1726824244,CONNECTED,SUCCESS,10.42.0.62,204.13.164.252,80,,,fd15:53b6:dead::2"),
			WantTimestamp:  "1726824244",
			WantState:      "CONNECTED",
			WantDesc:       "SUCCESS",
			WantLocalAddr:  "10.42.0.62",
			WantRemoteAddr: "204.13.164.252",
			WantRemotePort: "80",
		},
		{
			Input:          []byte("STATE:1726824244,CONNECTED,SUCCESS,10.42.0.62,204.13.164.252"),
			WantTimestamp:  "1726824244",
			WantState:      "CONNECTED",
			WantDesc:       "SUCCESS",
			WantLocalAddr:  "10.42.0.62",
			WantRemoteAddr: "204.13.164.252",
			WantRemotePort: "",
		},
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase.Input)

		var st *StateEvent
		var ok bool
		if st, ok = event.(*StateEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, st)
			continue
		}

		if got, want := st.RawTimestamp(), testCase.WantTimestamp; got != want {
			t.Errorf("test %d RawTimestamp returned %q; want %q", i, got, want)
		}
		if got, want := st.NewState(), testCase.WantState; got != want {
			t.Errorf("test %d NewState returned %q; want %q", i, got, want)
		}
		if got, want := st.Description(), testCase.WantDesc; got != want {
			t.Errorf("test %d Description returned %q; want %q", i, got, want)
		}
		if got, want := st.LocalTunnelAddr(), testCase.WantLocalAddr; got != want {
			t.Errorf("test %d LocalTunnelAddr returned %q; want %q", i, got, want)
		}
		if got, want := st.RemoteAddr(), testCase.WantRemoteAddr; got != want {
			t.Errorf("test %d RemoteAddr returned %q; want %q", i, got, want)
		}
		if got, want := st.RemotePort(), testCase.WantRemotePort; got != want {
			t.Errorf("test %d RemotePort returned %q; want %q", i, got, want)
		}
	}
}

func TestByteCountEvent(t *testing.T) {
	type TestCase struct {
		Input        []byte
		WantClientId string
		WantBytesIn  int
		WantBytesOut int
	}
	testCases := []TestCase{
		{
			Input:        []byte("BYTECOUNT:"),
			WantClientId: "",
			WantBytesIn:  0,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT:123,456"),
			WantClientId: "",
			WantBytesIn:  123,
			WantBytesOut: 456,
		},
		{
			Input:        []byte("BYTECOUNT:,"),
			WantClientId: "",
			WantBytesIn:  0,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT:5,"),
			WantClientId: "",
			WantBytesIn:  5,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT:,6"),
			WantClientId: "",
			WantBytesIn:  0,
			WantBytesOut: 6,
		},
		{
			Input:        []byte("BYTECOUNT:6"),
			WantClientId: "",
			WantBytesIn:  6,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT:wrong,bad"),
			WantClientId: "",
			WantBytesIn:  0,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT:1,2,3"),
			WantClientId: "",
			WantBytesIn:  1,
			WantBytesOut: 2,
		},
		{
			// Intentionally malformed BYTECOUNT event sent as BYTECOUNT_CLI
			Input:        []byte("BYTECOUNT_CLI:123,456"),
			WantClientId: "123",
			WantBytesIn:  456,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT_CLI:"),
			WantClientId: "",
			WantBytesIn:  0,
			WantBytesOut: 0,
		},
		{
			Input:        []byte("BYTECOUNT_CLI:abc123,123,456"),
			WantClientId: "abc123",
			WantBytesIn:  123,
			WantBytesOut: 456,
		},
		{
			Input:        []byte("BYTECOUNT_CLI:abc123,123"),
			WantClientId: "abc123",
			WantBytesIn:  123,
			WantBytesOut: 0,
		},
	}

	for i, testCase := range testCases {
		event := upgradeEvent(testCase.Input)

		var bc *ByteCountEvent
		var ok bool
		if bc, ok = event.(*ByteCountEvent); !ok {
			t.Errorf("test %d got %T; want %T", i, event, bc)
			continue
		}

		if got, want := bc.ClientId(), testCase.WantClientId; got != want {
			t.Errorf("test %d ClientId returned %q; want %q", i, got, want)
		}
		if got, want := bc.BytesIn(), testCase.WantBytesIn; got != want {
			t.Errorf("test %d BytesIn returned %d; want %d", i, got, want)
		}
		if got, want := bc.BytesOut(), testCase.WantBytesOut; got != want {
			t.Errorf("test %d BytesOut returned %d; want %d", i, got, want)
		}
	}
}
