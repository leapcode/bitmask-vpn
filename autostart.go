package main

type autostart interface {
	Disable() error
	Enable() error
}

type dummyAutostart struct{}

func (a *dummyAutostart) Disable() error {
	return nil
}

func (a *dummyAutostart) Enable() error {
	return nil
}
