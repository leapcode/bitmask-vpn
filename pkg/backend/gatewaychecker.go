package backend

import (
	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
)

// The gateway selector gets populated asynchronously, so this waits until the
// fetch has completed to update status. It also exits early if the backend is
// re-initialized (e.g. a provider switch closes the status channel).
func watchGateways(bm bitmask.Bitmask) {
	select {
	case <-bm.GetGatewaysFetchedCh():
		if len(bm.GetLocationQualityMap(bm.GetTransport())) > 0 {
			updateStatusForGateways()
		}
	case <-bm.GetStatusCloseCh():
	}
}

func updateStatusForGateways() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	go trigger(OnStatusChanged)
}
