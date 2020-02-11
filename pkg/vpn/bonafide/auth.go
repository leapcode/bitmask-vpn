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

package bonafide

type credentials struct {
	User     string
	Password string
}

// The authentication interface allows to get a Certificate in Pem format.
// We implement Anonymous Authentication (Riseup et al), and Sip (Libraries).

type authentication interface {
	needsCredentials() bool
	getToken(user, password string) ([]byte, error)
}
