package backend

import (
	"time"
)

// The gateway selector gets populated asynchronously, so this spawns a goroutine that
// checks whether they've been fetched to update status.
func (c connectionCtx) delayCheckForGateways() {
	go func() {
		cnt := 0
		for {
			if cnt > 60*2 {
				break
			}
			time.Sleep(time.Second * 5)
			transport := c.bm.GetTransport()
			locs := c.bm.ListLocationFullness(transport)
			if len(locs) != 0 {
				c.Locations = locs
				updateStatusForGateways()
				break
			}
		}
	}()
}

func updateStatusForGateways() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	go trigger(OnStatusChanged)
}
