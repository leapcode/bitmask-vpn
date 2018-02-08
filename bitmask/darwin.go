// +build darwin

package bitmask

import "os"

var ConfigPath = os.Getenv("HOME") + "/Library/Preferences/leap"
