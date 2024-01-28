package obfsvpn

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// ReadTCPFrameUDP reads from a tcp connection and returns a framed
// UDP buffer
func ReadTCPFrameUDP(tcpConn net.Conn, datagramBuffer []byte, lengthBuffer []byte) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered panic:", r)
		}
	}()
	var length16 uint16
	// Read the first 2 bytes from the tcp connection
	// These will be the length of the data
	_, err := io.ReadFull(tcpConn, lengthBuffer)
	if err != nil {
		return nil, fmt.Errorf("read err on %v from %v: %w", tcpConn.LocalAddr(), tcpConn.RemoteAddr(), err)
	}

	err = binary.Read(bytes.NewReader(lengthBuffer), binary.LittleEndian, &length16)
	if err != nil {
		return nil, fmt.Errorf("serialization error  %w", err)
	}

	readBuffer := datagramBuffer[:length16]
	_, err = io.ReadFull(tcpConn, readBuffer)
	if err != nil {
		return nil, fmt.Errorf("read err on %v from %v: %w", tcpConn.LocalAddr(), tcpConn.RemoteAddr(), err)
	}

	outSlice := make([]byte, len(readBuffer))
	copy(outSlice, readBuffer)
	return outSlice, nil
}

// ReadUDPFrameTCP reads from a udp connection and returns a framed
// TCP buffer
func ReadUDPFrameTCP(udpConn *net.UDPConn, datagramBuffer []byte) ([]byte, *net.UDPAddr, error) {
	var length16 uint16
	n, udpAddr, err := udpConn.ReadFromUDP(datagramBuffer)

	if err != nil {
		return nil, nil, fmt.Errorf("read err on %v: %w", udpConn.LocalAddr(), err)
	}

	readBuffer := datagramBuffer[:n]

	outSlice := make([]byte, len(readBuffer))
	copy(outSlice, readBuffer)

	length16 = uint16(n)
	lengthBuf := new(bytes.Buffer)
	err = binary.Write(lengthBuf, binary.LittleEndian, length16)
	if err != nil {
		return nil, nil, fmt.Errorf("serialization error  %w", err)
	}

	outSlice = append(lengthBuf.Bytes(), outSlice...)

	return outSlice, udpAddr, nil
}
