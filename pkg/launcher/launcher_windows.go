//go:build windows
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

package launcher

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf16"

	"0xacab.org/leap/bitmask-vpn/pkg/vpn/bonafide"
	"github.com/rs/zerolog/log"

	"github.com/natefinch/npipe"
)

const pipeName = `\\.\pipe\openvpn\service`

type Launcher struct {
	MngPass string
	Failed  bool
}

func NewLauncher() (*Launcher, error) {
	l := Launcher{}
	return &l, nil
}

func (l *Launcher) Close() error {
	return nil
}

func (l *Launcher) Check() (helpers bool, privilege bool, err error) {
	// TODO check if the named pipe exists
	log.Warn().Msg("bogus check on windows")
	return true, true, nil
}

func (l *Launcher) OpenvpnStart(flags ...string) error {
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
	log.Info().
		Str("args", args).
		Msg("Starting OpenVPN")

	timeout := 3 * time.Second
	conn, err := npipe.DialTimeout(pipeName, timeout)
	if err != nil {
		return fmt.Errorf("Could not open OpenVPN pipe. %v", err)

	}
	defer conn.Close()

	writeUTF16Bytes(&b, cwd)
	writeUTF16Bytes(&b, opts)
	writeUTF16Bytes(&b, `\n`)
	encoded := b.Bytes()

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	_, err = rw.Write(encoded)
	if err != nil {
		return fmt.Errorf("Could not write to OpenVPN pipe. %v", err)
	}
	rw.Flush()
	pid, err := getCommandResponse(rw)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get pid")
	}
	log.Debug().
		Int("pid", pid).
		Msg("OpenVPN is running")
	return nil
}

func (l *Launcher) OpenvpnStop() error {
	return nil
}

// TODO we will have to bring our helper back to do firewall

func (l *Launcher) FirewallStart(gateways []bonafide.Gateway) error {
	log.Warn().Msg("start: no firewall in windows")
	return nil
}

func (l *Launcher) FirewallStop() error {
	log.Debug().Msg("stop: no firewall in windows")
	return nil
}

func (l *Launcher) FirewallIsUp() bool {
	log.Debug().Msg("up: no firewall in windows")
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
