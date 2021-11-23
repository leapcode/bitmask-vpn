package main

import (
	"testing"
)

func TestGoodMotd(t *testing.T) {
	m, err := parseFile(defaultFile)
	if err != nil {
		t.Errorf("error parsing default file")
	}
	if m.Length() == 0 {
		t.Errorf("zero messages in file")
	}
	for _, msg := range m.Messages {
		if !msg.IsValid() {
			t.Errorf("invalid motd json at %s", defaultFile)
		}
	}
}

const emptyDate = `
{
    "motd": [{
        "begin":    "",
        "end":      "",
        "type":     "daily",
        "platform": "all",
        "urgency":  "normal",
        "text": [
          { "lang": "en",
            "str": "test"
	  }]
    }]
}`

func TestEmptyDateFails(t *testing.T) {
	m, err := parseJsonStr([]byte(emptyDate))
	if err != nil {
		t.Errorf("error parsing json")
	}
	if allValid(t, m) {
		t.Errorf("empty string should not be valid")
	}
}

const badEnd = `
{
    "motd": [{
	"begin":    "02 Jan 21 00:00 +0100",
	"end":      "01 Jan 21 00:00 +0100",
        "type":     "daily",
        "platform": "all",
        "urgency":  "normal",
        "text": [
          { "lang": "en",
            "str": "test"
	  }]
    }]
}`

func TestBadEnd(t *testing.T) {
	m, err := parseJsonStr([]byte(badEnd))
	if err != nil {
		t.Errorf("error parsing json")
	}
	if allValid(t, m) {
		t.Errorf("begin > end must fail")
	}
}

func allValid(t *testing.T, m Messages) bool {
	if m.Length() == 0 {
		t.Errorf("expected at least one message")

	}
	for _, msg := range m.Messages {
		if !msg.IsValid() {
			return false
		}
	}
	return true
}
