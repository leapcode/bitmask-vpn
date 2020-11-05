// +build !windows
// Copyright (C) 2018 LEAP
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

package bitmask

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	pmautostart "github.com/ProtonMail/go-autostart"
)

const (
	errorMsg = `An error has ocurred initializing the VPN: %v`
)

// Autostart holds the functions to enable and disable the application autostart
type Autostart interface {
	Disable() error
	Enable() error
}

// NewAutostart creates a handler for the autostart of your platform
func NewAutostart(appName string, iconPath string) Autostart {
	exec := os.Args
	if os.Getenv("SNAP") != "" {
		re := regexp.MustCompile("/snap/([^/]*)/")
		match := re.FindStringSubmatch(os.Args[0])
		if len(match) > 1 {
			snapName := match[1]
			exec = []string{fmt.Sprintf("/snap/bin/%s.launcher", snapName)}
		} else {
			log.Printf("Snap binary has unknown path: %v", os.Args[0])
		}
	}

	if exec[0][:2] == "./" || exec[0][:2] == ".\\" {
		var err error
		exec[0], err = filepath.Abs(exec[0])
		if err != nil {
			log.Printf("Error making the path absolute directory: %v", err)
		}
	}

	return &pmautostart.App{
		Name:        appName,
		Exec:        exec,
		DisplayName: appName,
		Icon:        iconPath,
	}
}

type DummyAutostart struct{}

func (a *DummyAutostart) Disable() error {
	return nil
}

func (a *DummyAutostart) Enable() error {
	return nil
}
