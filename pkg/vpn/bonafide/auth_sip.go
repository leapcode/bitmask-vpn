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
	certURI string
}

func (a *sipAuthentication) needsCredentials() bool {
	return true
}

func (a *sipAuthentication) getPemCertificate(cred *credentials) ([]byte, error) {
	if cred == nil {
		return nil, fmt.Errorf("Need bonafide credentials for sip auth")
	}
	token, err := a.getToken(cred)
	if err != nil {
		return nil, fmt.Errorf("Error while getting token: %s", err)
	}
	cert, err := a.getProtectedCert(a.certURI, string(token))
	if err != nil {
		return nil, fmt.Errorf("Error while getting cert: %s", err)
	}
	return cert, nil
}

func (a *sipAuthentication) getToken(cred *credentials) ([]byte, error) {
	/* TODO
	[ ] get token from disk?
	[ ] check if expired? set a goroutine to refresh it periodically?
	*/
	credJSON, err := formatCredentials(cred.User, cred.Password)
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

func (a *sipAuthentication) getProtectedCert(uri, token string) ([]byte, error) {
	req, err := http.NewRequest("POST", uri, strings.NewReader(""))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error while getting token: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error %d", resp.StatusCode)
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
