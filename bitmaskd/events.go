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
	"log"
	"net/http"
)

const (
	statusEvent = "VPN_STATUS_CHANGED"
)

func (b *Bitmask) eventsHandler() {
	b.send("events", "register", statusEvent)
	client := &http.Client{
		Timeout: 0,
	}
	for {
		resJSON, err := send(b.apiToken, client, "events", "poll")
		res, ok := resJSON.([]interface{})
		if err != nil || !ok || len(res) < 1 {
			continue
		}
		event, ok := res[0].(string)
		if !ok || event != statusEvent {
			continue
		}

		status, err := b.GetStatus()
		if err != nil {
			log.Printf("Error receiving status: %v", err)
			continue
		}
		b.statusCh <- status
	}
}
