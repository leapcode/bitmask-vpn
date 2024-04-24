package motd

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

const TimeString = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
const ExampleFile = "motd-example.json"

func ParseFile(f string) (Messages, error) {
	jsonFile, err := os.Open(f)
	if err != nil {
		return Messages{}, err
	}
	defer jsonFile.Close()
	byteVal, err := io.ReadAll(jsonFile)
	if err != nil {
		return Messages{}, err
	}
	return getFromJSON(byteVal)
}

func getFromJSON(b []byte) (Messages, error) {
	var m Messages
	err := json.Unmarshal(b, &m)
	return m, err
}

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

type LocalizedText struct {
	Lang string `json:"lang"`
	Str  string `json:"str"`
}

func (m *Message) IsValid() bool {
	valid := (m.IsValidBegin() && m.IsValidEnd() &&
		m.IsValidType() && m.IsValidPlatform() && m.IsValidUrgency() &&
		m.HasLocalizedText())
	return valid
}

func (m *Message) IsValidBegin() bool {
	_, err := time.Parse(TimeString, m.Begin)
	if err != nil {
		log.Warn().
			Err(err).
			Str("begin", m.Begin).
			Msg("Could not parse begin time in IsValidBegin")
		return false
	}
	return true
}

func (m *Message) IsValidEnd() bool {
	endTime, err := time.Parse(TimeString, m.End)
	if err != nil {
		log.Warn().
			Err(err).
			Str("end", m.End).
			Msg("Could not parse end time")
		return false
	}
	beginTime, err := time.Parse(TimeString, m.Begin)
	if err != nil {
		log.Warn().
			Err(err).
			Str("begin", m.Begin).
			Msg("Could not parse begin time")
		return false
	}
	if !beginTime.Before(endTime) {
		log.Warn().Msg("begin ts should be before end")
		return false
	}
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
