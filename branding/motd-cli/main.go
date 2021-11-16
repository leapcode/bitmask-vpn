package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const defaultFile = "motd-example.json"

const OK = "✓"
const WRONG = "☓"

/* TODO move structs to pkg/config/motd module, import from there */

type Messages struct {
	Messages []Message `json:"motd"`
}

func (m *Messages) Length() int {
	return len(m.Messages)
}

type Message struct {
	Begin    string          `json:"begin"`
	End      string          `json:"end"`
	Type     string          `json:"type"`
	Platform string          `json:"platform"`
	Urgency  string          `json:"urgency"`
	Text     []LocalizedText `json:"text"`
}

func (m *Message) IsValid() bool {
	valid := (m.IsValidBegin() && m.IsValidEnd() &&
		m.IsValidType() && m.IsValidPlatform() && m.IsValidUrgency() &&
		m.HasLocalizedText())
	return valid
}

func (m *Message) IsValidBegin() bool {
	// FIXME check that begin is before 1y for instance
	return true
}

func (m *Message) IsValidEnd() bool {
	// FIXME check end is within next year/months
	return true
}

func (m *Message) IsValidType() bool {
	switch m.Type {
	case "once", "daily":
		return true
	default:
		return false
	}
}

func (m *Message) IsValidPlatform() bool {
	switch m.Platform {
	case "windows", "linux", "osx", "all":
		return true
	default:
		return false
	}
}

func (m *Message) IsValidUrgency() bool {
	switch m.Urgency {
	case "normal", "critical":
		return true
	default:
		return false
	}
}

func (m *Message) HasLocalizedText() bool {
	return len(m.Text) > 0
}

type LocalizedText struct {
	Lang string `json:"lang"`
	Str  string `json:"str"`
}

func main() {
	file := flag.String("file", "", "file to validate")
	url := flag.String("url", "", "url to validate")
	flag.Parse()

	f := *file
	u := *url

	if u != "" {
		fmt.Println("url:", u)
		f = downloadToTempFile(u)
	} else {
		if f == "" {
			f = defaultFile
		}
		fmt.Println("file:", f)
	}
	m := parseFile(f)
	fmt.Printf("count: %v\n", m.Length())
	fmt.Println()
	for i, msg := range m.Messages {
		fmt.Printf("Message %d %v\n-----------\n", i+1, mark(msg.IsValid()))
		fmt.Printf("Type: %s %v\n", msg.Type, mark(msg.IsValidType()))
		fmt.Printf("Platform: %s %v\n", msg.Platform, mark(msg.IsValidPlatform()))
		fmt.Printf("Urgency: %s %v\n", msg.Urgency, mark(msg.IsValidUrgency()))
		fmt.Printf("Languages: %d %v\n", len(msg.Text), mark(msg.HasLocalizedText()))
		if !msg.IsValid() {
			os.Exit(1)
		}
	}
}

func downloadToTempFile(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	out, err := ioutil.TempFile("/tmp/", "motd-linter")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, _ = io.Copy(out, resp.Body)
	fmt.Println("File downloaded to", out.Name())
	return out.Name()
}

func parseFile(f string) Messages {
	jsonFile, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()
	byteVal, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}
	var m Messages
	json.Unmarshal(byteVal, &m)
	return m
}

func mark(val bool) string {
	if val {
		return OK
	} else {
		return WRONG
	}
}
