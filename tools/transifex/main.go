package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"golang.org/x/text/message/pipeline"
)

const (
	outGotext      = "out.gotext.json"
	messagesGotext = "messages.gotext.json"
)

type transifex map[string]string

func main() {
	if len(os.Args) < 2 {
		panic("g2t or t2g should be passed as argument")
	}

	switch os.Args[1] {
	case "g2t":
		g2t(func(m pipeline.Message) string { return m.Message.Msg })
	case "lang2t":
		g2t(func(m pipeline.Message) string { return m.Translation.Msg })
	case "t2g":
		t2g()
	default:
		panic("g2t or t2g should be passed as argument")
	}
}

func g2t(getMessage func(pipeline.Message) string) {
	if len(os.Args) < 4 {
		panic(fmt.Sprintf("usage: %s g2t inFile outFile", os.Args[0]))
	}

	inF, err := os.Open(os.Args[2])
	if err != nil {
		panic(fmt.Sprintf("Can't open input file %s: %v", os.Args[2], err))
	}
	outF, err := os.Create(os.Args[3])
	if err != nil {
		panic(fmt.Sprintf("Can't open output file %s: %v", os.Args[3], err))
	}

	toTransifex(inF, outF, getMessage)
}

func t2g() {
	if len(os.Args) < 3 {
		panic(fmt.Sprintf("usage: %s t2g localeFolder", os.Args[0]))
	}

	origF, err := os.Open(path.Join(os.Args[2], outGotext))
	if err != nil {
		panic(fmt.Sprintf("Can't open file %s/%s: %v", os.Args[3], outGotext, err))
	}
	outF, err := os.Create(path.Join(os.Args[2], messagesGotext))
	if err != nil {
		panic(fmt.Sprintf("Can't open output file %s/%v: %v", os.Args[3], messagesGotext, err))
	}
	toGotext(origF, os.Stdin, outF)
}

func toTransifex(inF, outF *os.File, getMessage func(pipeline.Message) string) {
	messages := pipeline.Messages{}
	dec := json.NewDecoder(inF)
	err := dec.Decode(&messages)
	if err != nil {
		panic(fmt.Sprintf("An error ocurred decoding json: %v", err))
	}

	transfx := make(transifex)
	for _, m := range messages.Messages {
		transfx[m.ID[0]] = getMessage(m)
	}
	enc := json.NewEncoder(outF)
	enc.SetIndent("", "    ")
	err = enc.Encode(transfx)
	if err != nil {
		panic(fmt.Sprintf("An error ocurred encoding json: %v", err))
	}
}

func toGotext(origF, inF, outF *os.File) {
	transfx := make(transifex)
	dec := json.NewDecoder(inF)
	err := dec.Decode(&transfx)
	if err != nil {
		panic(fmt.Sprintf("An error ocurred decoding json: %v", err))
	}

	messages := pipeline.Messages{}
	dec = json.NewDecoder(origF)
	err = dec.Decode(&messages)
	if err != nil {
		panic(fmt.Sprintf("An error ocurred decoding orig json: %v", err))
	}

	for k, v := range transfx {
		found := false
		for i, m := range messages.Messages {
			if m.ID[0] == k {
				messages.Messages[i].Translation.Msg = v
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("The original document doesn't have id: %s\n", k)
		}
	}

	enc := json.NewEncoder(outF)
	enc.SetIndent("", "    ")
	err = enc.Encode(messages)
	if err != nil {
		panic(fmt.Sprintf("An error ocurred encoding json: %v", err))
	}
}
