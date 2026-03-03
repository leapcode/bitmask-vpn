package backend

import (
	"time"
)

// The gateway selector gets populated asynchronously, so this spawns a goroutine that
// checks whether they've been fetched to update status.
func (c connectionCtx) delayCheckForGateways() {
	go func() {
		for cnt := 0; cnt <= 60*2; cnt++ {
			time.Sleep(time.Second * 5)
			transport := c.bm.GetTransport()
			locs := c.bm.GetLocationQualityMap(transport)
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
