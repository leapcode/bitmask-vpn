// Package demux implements low-level demultiplexing of the stream of
// messages sent from OpenVPN on the management channel.
//
// OpenVPN's protocol includes two different kinds of message from the OpenVPN
// process: replies to commands sent by the management client, and asynchronous
// event notifications.
//
// This package's purpose is to split these messages into two separate streams,
// so that functions executing command/response sequences can just block
// on the reply channel while an event loop elsewhere deals with any async
// events that might show up.
package demux
