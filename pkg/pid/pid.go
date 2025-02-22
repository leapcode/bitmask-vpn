// Copyright (C) 2018-2020 LEAP
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

package pid

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/keybase/go-ps"
)

var (
	pidFile           = filepath.Join(config.Path, "systray.pid")
	AlreadyRunningErr = errors.New("another instance is already running")
)

func AcquirePID() error {
	pid := syscall.Getpid()
	current, err := getPID()
	if err != nil {
		log.Err(err).Msg("Error reading pid file")
	}

	if current != pid && pidRunning(current) {
		return fmt.Errorf("%w, pid: %d", AlreadyRunningErr, current)
	}

	return setPID(pid)
}

func ReleasePID() error {
	pid := syscall.Getpid()
	current, err := getPID()
	if err != nil {
		return err
	}
	if current != 0 && current != pid {
		return fmt.Errorf("Can't release pid file, is not own by this process")
	}

	if current == pid {
		return os.Remove(pidFile)
	}
	return nil
}

func getPID() (int, error) {
	_, err := os.Stat(pidFile)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	file, err := os.Open(pidFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return strconv.Atoi(string(b))
}

func setPID(pid int) error {
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	return err
}

func pidRunning(pid int) bool {
	if pid == 0 {
		return false
	}
	proc, err := ps.FindProcess(pid)
	if err != nil {
		log.Warn().
			Err(err).
			Int("pid", pid).
			Msg("Could not find running process")
		return false
	}
	if proc == nil {
		return false
	}
	log.Debug().
		Int("pid", pid).
		Str("executable", proc.Executable()).
		Msg("Found a running process with")
	return strings.Contains(os.Args[0], proc.Executable())
}
