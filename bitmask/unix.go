// +build !windows,!darwin

package bitmask

import "os"

var configPath = os.Getenv("HOME") + "/.config/leap"
