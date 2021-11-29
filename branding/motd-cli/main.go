package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"0xacab.org/leap/bitmask-vpn/pkg/motd"
)

const OK = "✓"
const WRONG = "☓"

func main() {
	file := flag.String("file", "", "file to validate")
	url := flag.String("url", "", "url to validate")
	flag.Parse()

	f := *file
	u := *url

	var m motd.Messages
	var err error

	if u != "" {
		fmt.Println("url:", u)
		f = downloadToTempFile(u)
	} else {
		if f == "" {
			f = filepath.Join("../../pkg/motd/", motd.ExampleFile)
		}
		fmt.Println("file:", f)
	}
	m, err = motd.ParseFile(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("count: %v\n", m.Length())
	fmt.Println()
	for i, msg := range m.Messages {
		fmt.Printf("Message %d %v\n-----------\n", i+1, mark(msg.IsValid()))
		fmt.Printf("Type: %s %v\n", msg.Type, mark(msg.IsValidType()))
		fmt.Printf("Platform: %s %v\n", msg.Platform, mark(msg.IsValidPlatform()))
		fmt.Printf("Urgency: %s %v\n", msg.Urgency, mark(msg.IsValidUrgency()))
		fmt.Printf("Languages: %d %v\n", len(msg.Text), mark(msg.HasLocalizedText()))
		for _, t := range msg.Text {
			fmt.Printf(t.Str)
		}
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

func mark(val bool) string {
	if val {
		return OK
	} else {
		return WRONG
	}
}
