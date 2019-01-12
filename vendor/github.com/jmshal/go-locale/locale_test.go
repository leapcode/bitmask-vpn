package go_locale

import "testing"

func TestLocale(t *testing.T) {
	if lc, err := DetectLocale(); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("detected locale: %q", lc)
	}
}
