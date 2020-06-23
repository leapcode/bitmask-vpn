package bitmask

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

/* functions for local authentication of control endpoints */

const tokenPath = "/dev/shm/bitmask-token"

func GenerateAuthToken() {
	if runtime.GOOS != "linux" {
		log.Println("Authentication token only implemented in linux at the moment.")
		return
	}
	t := getRandomString()
	err := ioutil.WriteFile(tokenPath, []byte(t), os.FileMode(int(0600)))
	if err != nil {
		log.Println("Could not write authentication token.")
	}
}

func ReadAuthToken() string {
	if runtime.GOOS != "linux" {
		log.Println("Authentication token only implemented in linux at the moment.")
		return ""
	}
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		log.Println("Error reading token:", err)
	}
	return string(token)
}

func getRandomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 40
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
