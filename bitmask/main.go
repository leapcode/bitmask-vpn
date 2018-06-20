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

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"
)

const (
	timeout    = time.Second * 15
	url        = "http://localhost:7070/API/"
	headerAuth = "X-Bitmask-Auth"
)

// Bitmask holds the bitmask client data
type Bitmask struct {
	client   *http.Client
	apiToken string
	statusCh chan string
}

// Init the connection to bitmask
func Init() (*Bitmask, error) {
	statusCh := make(chan string)
	client := &http.Client{
		Timeout: timeout,
	}

	err := waitForBitmaskd()
	if err != nil {
		return nil, err
	}

	apiToken, err := getToken()
	if err != nil {
		return nil, err
	}

	b := Bitmask{client, apiToken, statusCh}
	go b.eventsHandler()
	return &b, nil
}

// GetStatusCh returns a channel that will recieve VPN status changes
func (b *Bitmask) GetStatusCh() chan string {
	return b.statusCh
}

// Close the connection to bitmask
func (b *Bitmask) Close() {
	_, err := b.send("core", "stop")
	if err != nil {
		log.Printf("Got an error stopping bitmaskd: %v", err)
	}
}

// Version gets the bitmask version string
func (b *Bitmask) Version() (string, error) {
	res, err := b.send("core", "version")
	if err != nil {
		return "", err
	}
	return res["version_core"].(string), nil
}

func waitForBitmaskd() error {
	var err error
	for i := 0; i < 30; i++ {
		resp, err := http.Post(url, "", nil)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		log.Printf("Bitmask is not ready (iteration %d): %v", i, err)
		time.Sleep(1 * time.Second)
	}
	return err
}

func (b *Bitmask) send(parts ...interface{}) (map[string]interface{}, error) {
	resJSON, err := send(b.apiToken, b.client, parts...)
	if err != nil {
		return nil, err
	}
	result, ok := resJSON.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not valid response")
	}
	return result, nil
}

func send(apiToken string, client *http.Client, parts ...interface{}) (interface{}, error) {
	apiSection, _ := parts[0].(string)
	reqBody, err := json.Marshal(parts[1:])
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url+apiSection, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Add(headerAuth, apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	resJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseResponse(resJSON)
}

func parseResponse(resJSON []byte) (interface{}, error) {
	var response struct {
		Result interface{}
		Error  string
	}
	err := json.Unmarshal(resJSON, &response)
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	return response.Result, err
}

func getToken() (string, error) {
	var err error
	path := path.Join(ConfigPath, "authtoken")
	for i := 0; i < 30; i++ {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			return string(b), nil
		}
		log.Printf("Auth token is not ready (iteration %d): %v", i, err)
		time.Sleep(1 * time.Second)
	}
	return "", err
}
