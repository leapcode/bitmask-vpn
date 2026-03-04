package backend

import "testing"

func TestInitOptsFromJSON_NoProvider(t *testing.T) {
	// Minimal providers JSON with an empty default field.
	providersJSON := `{"default":"","providers":[{"name":"bitmask","applicationName":"Bitmask","binaryName":"bitmask-vpn"}]}`
	opts := InitOptsFromJSON("", providersJSON)
	if opts.ProviderOptions != nil {
		t.Fatalf("expected ProviderOptions to be nil for empty provider, got %+v", opts.ProviderOptions)
	}
	// Ensure that the list of AvailableProviders still contains the known provider.
	found := false
	for _, p := range opts.AvailableProviders {
		if p == "bitmask" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected AvailableProviders to include 'bitmask'")
	}
}
