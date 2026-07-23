package backend

import "testing"

func TestInitOptsFromJSON_NoProvider(t *testing.T) {
	// Minimal providers JSON with an empty default field. The "provider" field
	// is the canonical identifier (INI section header); "name" is the display
	// name shown in the UI.
	providersJSON := `{"default":"","providers":[{"provider":"bitmask","name":"Bitmask","applicationName":"Bitmask","binaryName":"bitmask-vpn"}]}`
	opts := InitOptsFromJSON("", providersJSON)
	if opts.ProviderOptions != nil {
		t.Fatalf("expected ProviderOptions to be nil for empty provider, got %+v", opts.ProviderOptions)
	}
	// Ensure that the list of AvailableProviders still contains the known provider.
	found := false
	for _, p := range opts.AvailableProviders {
		if p.ID == "bitmask" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected AvailableProviders to include 'bitmask'")
	}
}

// TestInitOptsFromJSON_ProviderIDMatches verifies that provider selection is by
// the canonical id (the "provider" field / INI section header), not the display
// name. This covers the coopvpn/CoopVPN case-mismatch regression: the id is
// "coopvpn" while the display name is "CoopVPN", and matching must succeed
// using the id.
func TestInitOptsFromJSON_ProviderIDMatches(t *testing.T) {
	// Reset the cached providers so this test parses its own JSON.
	providers = nil

	providersJSON := `{"default":"coopvpn","providers":[
		{"provider":"riseup","name":"riseup","applicationName":"RiseupVPN","binaryName":"riseup-vpn"},
		{"provider":"coopvpn","name":"CoopVPN","applicationName":"Bitmask","binaryName":"bitmask-vpn"}
	]}`
	opts := InitOptsFromJSON("coopvpn", providersJSON)
	if opts.ProviderOptions == nil {
		t.Fatalf("expected ProviderOptions to be set for provider 'coopvpn'")
	}
	if opts.ProviderOptions.Provider != "coopvpn" {
		t.Fatalf("expected Provider 'coopvpn', got %q", opts.ProviderOptions.Provider)
	}
	if opts.ProviderOptions.DisplayName != "CoopVPN" {
		t.Fatalf("expected DisplayName 'CoopVPN', got %q", opts.ProviderOptions.DisplayName)
	}
	if opts.ProviderOptions.AppName != "Bitmask" {
		t.Fatalf("expected AppName 'Bitmask', got %q", opts.ProviderOptions.AppName)
	}
}
