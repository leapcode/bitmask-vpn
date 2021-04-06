// Copyright (c) 2016 Martin Atkins

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

package demux

import (
	"bufio"
	"io"
)

var readErrSynthEvent = []byte("FATAL:Error reading from OpenVPN")

// Demultiplex reads from the given io.Reader, assumed to be the client
// end of an OpenVPN Management Protocol connection, and splits it into
// distinct messages from OpenVPN.
//
// It then writes the raw message buffers into either replyCh or eventCh
// depending on whether each message is a reply to a client command or
// an asynchronous event notification.
//
// The buffers written to replyCh are entire raw message lines (without the
// trailing newlines), while the buffers written to eventCh are the raw
// event strings with the prototcol's leading '>' indicator omitted.
//
// The caller should usually provide buffered channels of sufficient buffer
// depth so that the reply channel will not be starved by slow event
// processing.
//
// Once the io.Reader signals EOF, eventCh will be closed, then replyCh
// will be closed, and then this function will return.
//
// As a special case, if a non-EOF error occurs while reading from the
// io.Reader then a synthetic "FATAL" event will be written to eventCh
// before the two buffers are closed and the function returns. This
// synthetic message will have the error message "Error reading from OpenVPN".
func Demultiplex(r io.Reader, replyCh, eventCh chan<- []byte) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		buf := scanner.Bytes()

		if len(buf) < 1 {
			// Should never happen but we'll be robust and ignore this,
			// rather than crashing below.
			continue
		}

		// Asynchronous messages always start with > to differentiate
		// them from replies.
		if buf[0] == '>' {
			// Trim off the > when we post the message, since it's
			// redundant after we've demuxed.
			eventCh <- buf[1:]
		} else {
			replyCh <- buf
		}
	}

	if err := scanner.Err(); err != nil {
		// Generate a synthetic FATAL event so that the caller can
		// see that the connection was not gracefully closed.
		eventCh <- readErrSynthEvent
	}

	close(eventCh)
	close(replyCh)
}
