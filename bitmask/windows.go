// +build windows

package bitmask

import "os"

var configPath = os.Getenv("APPDATA") + "\\leap"
