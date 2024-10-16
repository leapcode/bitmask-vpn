// models contains models that relate to client-only concepts, like objects we keep in the internal storage.
// For models coming from the menshen API Spec, see ../../models
package models

import "time"

// TODO - can simplify into a single model. URL and Raw are equivalent to "Blob"
// TODO - introduce "type" field.

// Introducer keeps metadata about introducers that the user has added to the Bitmask application. Introducers are expected to be transmitted off-band.
type Introducer struct {
	FQDN string
	// URL is the canonical URL. It should be stored after validation and writing in the canonical order, since
	// we will check for uniqueness.
	URL       string
	CreatedAt time.Time
	LastUsed  time.Time
}

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
