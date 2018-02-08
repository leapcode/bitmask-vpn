// +build !windows,!darwin

package bitmask

import "os"

var ConfigPath = os.Getenv("HOME") + "/.config/leap"
