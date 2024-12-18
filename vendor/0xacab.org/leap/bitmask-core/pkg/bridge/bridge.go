package bridge

import "time"

// Bridge is a private bridge.
type Bridge struct {
	Name     string
	Location string
	Type     string
	// Raw is the raw JSON serialization of the bridge. We could also use the menshen model as a nested struct.
	Raw       string
	CreatedAt time.Time
	LastUsed  time.Time
}
