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
	"log"
	"net/http"
	"strings"
)

type SipAuthentication struct {
	bonafide *Bonafide
}

func (a *SipAuthentication) GetPemCertificate() ([]byte, error) {
	cred := a.bonafide.credentials
	if cred == nil {
		return nil, fmt.Errorf("Need bonafide credentials for sip auth")
	}
	credJson, err := formatCredentials(cred.User, cred.Password)
	if err != nil {
		return nil, fmt.Errorf("Cannot encode credentials: %s", err)
	}
	token, err := a.getToken(credJson)
	if err != nil {
		return nil, fmt.Errorf("Error while getting token: %s", err)
	}
	cert, err := a.getProtectedCert(string(token))
	if err != nil {
		return nil, fmt.Errorf("Error while getting cert: %s", err)
	}
	return cert, nil
}

func (a *SipAuthentication) getProtectedCert(token string) ([]byte, error) {
	req, err := http.NewRequest("POST", certAPI, strings.NewReader(""))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := a.bonafide.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error while getting token: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Cannot get cert: Error %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func (a *SipAuthentication) getToken(credJson string) ([]byte, error) {
	/* TODO
	[ ] get token from disk?
	[ ] check if expired? set a goroutine to refresh it periodically?
	*/
	resp, err := http.Post(authAPI, "text/json", strings.NewReader(credJson))
	if err != nil {
		log.Fatal("Error on auth request: ", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Cannot get token: Error %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func formatCredentials(user, pass string) (string, error) {
	c := Credentials{User: user, Password: pass}
	credJson, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(credJson), nil
}
