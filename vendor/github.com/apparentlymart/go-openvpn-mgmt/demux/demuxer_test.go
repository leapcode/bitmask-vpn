package demux

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestDemultiplex(t *testing.T) {
	type TestCase struct {
		Input           []string
		ExpectedReplies []string
		ExpectedEvents  []string
	}

	testCases := []TestCase{
		{
			Input:           []string{},
			ExpectedReplies: []string{},
			ExpectedEvents:  []string{},
		},
		{
			Input: []string{
				"SUCCESS: foo bar baz",
			},
			ExpectedReplies: []string{
				"SUCCESS: foo bar baz",
			},
			ExpectedEvents: []string{},
		},
		{
			Input: []string{
				">STATE:1234,ASSIGN_IP,,10.0.0.1,",
			},
			ExpectedReplies: []string{},
			ExpectedEvents: []string{
				"STATE:1234,ASSIGN_IP,,10.0.0.1,",
			},
		},
		{
			Input: []string{
				">STATE:1234,ASSIGN_IP,,10.0.0.1,",
				">STATE:5678,ASSIGN_IP,,10.0.0.1,",
				">STATE:9012,ASSIGN_IP,,10.0.0.1,",
			},
			ExpectedReplies: []string{},
			ExpectedEvents: []string{
				"STATE:1234,ASSIGN_IP,,10.0.0.1,",
				"STATE:5678,ASSIGN_IP,,10.0.0.1,",
				"STATE:9012,ASSIGN_IP,,10.0.0.1,",
			},
		},
		{
			Input: []string{
				">STATE:1234,ASSIGN_IP,,10.0.0.1,",
				"SUCCESS: foo bar baz",
				">STATE:5678,ASSIGN_IP,,10.0.0.1,",
			},
			ExpectedReplies: []string{
				"SUCCESS: foo bar baz",
			},
			ExpectedEvents: []string{
				"STATE:1234,ASSIGN_IP,,10.0.0.1,",
				"STATE:5678,ASSIGN_IP,,10.0.0.1,",
			},
		},
		{
			Input: []string{
				"SUCCESS: foo bar baz",
				">STATE:1234,ASSIGN_IP,,10.0.0.1,",
				"SUCCESS: baz bar foo",
			},
			ExpectedReplies: []string{
				"SUCCESS: foo bar baz",
				"SUCCESS: baz bar foo",
			},
			ExpectedEvents: []string{
				"STATE:1234,ASSIGN_IP,,10.0.0.1,",
			},
		},
	}

	for i, testCase := range testCases {
		r := mockReader(testCase.Input)
		gotReplies, gotEvents := captureMsgs(r)

		if !reflect.DeepEqual(gotReplies, testCase.ExpectedReplies) {
			t.Errorf(
				"test %d returned incorrect replies\ngot  %#v\nwant %#v",
				i, gotReplies, testCase.ExpectedReplies,
			)
		}

		if !reflect.DeepEqual(gotEvents, testCase.ExpectedEvents) {
			t.Errorf(
				"test %d returned incorrect events\ngot  %#v\nwant %#v",
				i, gotEvents, testCase.ExpectedEvents,
			)
		}
	}
}

func TestDemultiplex_error(t *testing.T) {
	r := &alwaysErroringReader{}

	gotReplies, gotEvents := captureMsgs(r)

	expectedReplies := []string{}
	expectedEvents := []string{
		"FATAL:Error reading from OpenVPN",
	}

	if !reflect.DeepEqual(gotReplies, expectedReplies) {
		t.Errorf(
			"incorrect replies\ngot  %#v\nwant %#v",
			gotReplies, expectedReplies,
		)
	}

	if !reflect.DeepEqual(gotEvents, expectedEvents) {
		t.Errorf(
			"incorrect events\ngot  %#v\nwant %#v",
			gotEvents, expectedEvents,
		)
	}
}

func mockReader(msgs []string) io.Reader {
	var buf []byte
	for _, msg := range msgs {
		buf = append(buf, []byte(msg)...)
		buf = append(buf, '\n')
	}
	return bytes.NewReader(buf)
}

func captureMsgs(r io.Reader) (replies, events []string) {
	replyCh := make(chan []byte)
	eventCh := make(chan []byte)

	replies = make([]string, 0)
	events = make([]string, 0)

	go Demultiplex(r, replyCh, eventCh)

	for replyCh != nil || eventCh != nil {
		select {

		case msg, ok := <-replyCh:
			if ok {
				replies = append(replies, string(msg))
			} else {
				replyCh = nil
			}

		case msg, ok := <-eventCh:
			if ok {
				events = append(events, string(msg))
			} else {
				eventCh = nil
			}

		}

	}

	return replies, events
}

type alwaysErroringReader struct{}

func (r *alwaysErroringReader) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("mock error")
}

// Somewhat-contrived example of blocking for a reply while concurrently
// processing asynchronous events.
func ExampleDemultiplex() {
	// In a real caller we would have a net.IPConn as our reader,
	// but we'll use a bytes reader here as a test.
	r := bytes.NewReader([]byte(
		// A reply to a hypothetical command interspersed between
		// two asynchronous events.
		">HOLD:Waiting for hold release\nSUCCESS: foo\n>FATAL:baz\n",
	))

	// No strong need for buffering on this channel because usually
	// a message sender will immediately block waiting for the
	// associated response message.
	replyCh := make(chan []byte)

	// Make sure the event channel buffer is deep enough that slow event
	// processing won't significantly delay synchronous replies. If you
	// process events quickly, or if you aren't sending any commands
	// concurrently with acting on events, then this is not so important.
	eventCh := make(chan []byte, 10)

	// Start demultiplexing the message stream in the background.
	// This goroutine will exit once the reader signals EOF.
	go Demultiplex(r, replyCh, eventCh)

	// Some coroutine has sent a hypothetical message to OpenVPN,
	// and it can directly block until the associated reply arrives.
	// The events will be concurrently handled by our event loop
	// below while we wait for the reply to show up.
	go func() {
		replyMsgBuf := <-replyCh
		fmt.Printf("Command reply: %s\n", string(replyMsgBuf))
	}()

	// Main event loop deals with the async events as they arrive,
	// independently of any commands that are pending.
	for msgBuf := range eventCh {
		fmt.Printf("Event: %s\n", string(msgBuf))
	}
}
