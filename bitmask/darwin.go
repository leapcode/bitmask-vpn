// +build darwin

package bitmask

import "os"

var configPath = os.Getenv("HOME") + "/Library/Preferences/leap"
