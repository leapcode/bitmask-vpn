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
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-vpn/pkg/config"
)

type sipAuthentication struct {
	client  httpClient
	authURI string
}

func (a *sipAuthentication) needsCredentials() bool {
	return true
}

func (a *sipAuthentication) getToken(user, password string) ([]byte, error) {
	/* TODO refresh session token periodically */
	if hasRecentToken() {
		return readToken()
	}
	credJSON, err := formatCredentials(user, password)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not encode credentials")
		return nil, fmt.Errorf("TokenErrBadCred")
	}
	resp, err := a.client.Post(a.authURI, "text/json", strings.NewReader(credJSON))
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Failed auth request")
		if os.IsTimeout(err) {
			return nil, fmt.Errorf("TokenErrTimeout")
		}
		return nil, fmt.Errorf("TokenErrBadPost")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("TokenErrBadStatus %d", resp.StatusCode)
	}
	token, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	writeToken(token)
	return token, nil
}

func getTokenPath() string {
	return path.Join(config.Path, config.ProviderConfig.ApplicationName+".token")
}

func writeToken(token []byte) {
	tp := getTokenPath()
	err := ioutil.WriteFile(tp, token, 0600)
	if err != nil {
		log.Warn().
			Err(err).
			Str("tokenPath", tp).
			Msg("Could not write token")
	}
}

func readToken() ([]byte, error) {
	f, err := os.Open(getTokenPath())
	if err != nil {
		log.Warn().
			Err(err).
			Str("tokenPath", getTokenPath()).
			Msg("Could not open token file")
		return nil, err
	}
	token, err := ioutil.ReadAll(f)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not read token")
		return nil, err
	}
	return token, nil
}

func hasRecentToken() bool {
	/* See https://0xacab.org/leap/bitmask-vpn/-/issues/346 for ability to refresh tokens,
	   when implemented that should be tried in a goroutine */
	statinfo, err := os.Stat(getTokenPath())
	if err != nil {
		return false
	}
	lastWrote := statinfo.ModTime().Unix()
	/* in vpnweb we set the duration of the token to 24 hours */
	old := time.Now().Add(-time.Hour * 20).Unix()
	return lastWrote >= old
}

func formatCredentials(user, pass string) (string, error) {
	c := credentials{User: user, Password: pass}
	credJSON, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(credJSON), nil
}
