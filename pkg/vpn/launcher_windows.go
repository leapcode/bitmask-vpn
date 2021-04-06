// +build windows
// Copyright (C) 2018-2021 LEAP
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package vpn

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"github.com/natefinch/npipe"
)

const pipeName = `\\.\pipe\openvpn\service`

type launcher struct {
	mngPass string
}

func newLauncher() (*launcher, error) {
	l := launcher{}
	return &l, nil
}

func (l *launcher) close() error {
	return nil
}

func (l *launcher) check() (helpers bool, privilege bool, err error) {
	// TODO check if the named pipe exists
	log.Println("bogus check on windows")
	return true, true, nil
}

func (l *launcher) openvpnStart(flags ...string) error {
	var b bytes.Buffer
	/* DELETE-ME
	var filtered []string
	for _, v := range flags {
		if v != "--tun-ipv6" {
			filtered = append(filtered, v)
		}
	}
	*/

	cwd, _ := os.Getwd()
	opts := `--client --dev tun --block-outside-dns --redirect-gateway --script-security 0 ` + strings.Join(flags, " ")
	log.Println("openvpn start: ", opts)

	timeout := 3 * time.Second
	conn, err := npipe.DialTimeout(pipeName, timeout)
	if err != nil {
		fmt.Println("ERROR opening pipe")
		return errors.New("cannot open openvpn pipe")

	}
	defer conn.Close()

	writeUTF16Bytes(&b, cwd)
	writeUTF16Bytes(&b, opts)
	writeUTF16Bytes(&b, `\n`)
	encoded := b.Bytes()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	_, err = rw.Write(encoded)
	if err != nil {
		log.Println("ERROR writing to pipe")
		return errors.New("cannot write to openvpn pipe")
	}
	rw.Flush()
	pid, err := getCommandResponse(rw)
	if err != nil {
		log.Println("ERROR getting pid")
	}
	log.Println("OpenVPN PID:", pid)
	return nil
}

func (l *launcher) openvpnStop() error {
	return nil
}

// TODO we will have to bring our helper back to do firewall

func (l *launcher) firewallStart(gateways []bonafide.Gateway) error {
	log.Println("start: no firewall in windows")
	return nil
}

func (l *launcher) firewallStop() error {
	log.Println("stop: no firewall in windows")
	return nil
}

func (l *launcher) firewallIsUp() bool {
	log.Println("up: no firewall in windows")
	return false
}

func writeUTF16Bytes(b *bytes.Buffer, in string) {
	var u16 []uint16 = utf16.Encode([]rune(in + "\x00"))
	binary.Write(b, binary.LittleEndian, u16)
}

func decodeUTF16String(s string) int {
	var code int
	var dec []byte
	for _, v := range []byte(s) {
		if byte(v) != byte(0) {
			dec = append(dec, v)
		}
	}
	_, err := fmt.Sscanf(string(dec), "%v", &code)
	if err != nil {
		fmt.Println("ERROR decoding")
	}
	return code
}

func getCommandResponse(rw *bufio.ReadWriter) (int, error) {
	msg, err := rw.ReadString('\n')
	if err != nil {
		fmt.Println("ERROR reading")
	}
	ok := decodeUTF16String(msg)
	if ok != 0 {
		return -1, errors.New("command failed")
	}
	msg, err = rw.ReadString('\n')
	if err != nil {
		fmt.Println("ERROR reading")
	}
	pid := decodeUTF16String(msg)
	if pid == 0 {
		return -1, errors.New("command failed")
	}
	return pid, nil
}
