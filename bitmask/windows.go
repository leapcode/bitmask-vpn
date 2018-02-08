// +build windows

package bitmask

import "os"

var ConfigPath = os.Getenv("APPDATA") + "\\leap"
