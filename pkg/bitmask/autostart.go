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

// Autostart holds the functions to enable and disable the application autostart
type Autostart interface {
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
