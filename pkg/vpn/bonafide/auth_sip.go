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

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type sipAuthentication struct {
	client  httpClient
	authURI string
}

func (a *sipAuthentication) needsCredentials() bool {
	return true
}

func (a *sipAuthentication) getToken(user, password string) ([]byte, error) {
	/* TODO
	[ ] get token from disk?
	[ ] check if expired? set a goroutine to refresh it periodically?
	*/
	credJSON, err := formatCredentials(user, password)
	if err != nil {
		return nil, fmt.Errorf("Cannot encode credentials: %s", err)
	}
	resp, err := http.Post(a.authURI, "text/json", strings.NewReader(credJSON))
	if err != nil {
		return nil, fmt.Errorf("Error on auth request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Cannot get token: Error %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func formatCredentials(user, pass string) (string, error) {
	c := credentials{User: user, Password: pass}
	credJSON, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(credJSON), nil
}
