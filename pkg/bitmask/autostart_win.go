//go:build windows
// +build windows

package bitmask

// Workaround for broken autostart package on windows.

type DummyAutostart struct{}

func (a *DummyAutostart) Disable() error {
	return nil
}

func (a *DummyAutostart) Enable() error {
	return nil
}

type Autostart interface {
	Disable() error
	Enable() error
}

func NewAutostart(appName string, iconPath string) Autostart {
	return &DummyAutostart{}
}
