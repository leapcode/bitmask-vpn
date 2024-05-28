package helper

import (
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
)

func getFirstAvailablePortFrom(port int) int {
	for {
		if isPortAvailable(port) {
			return port
		}
		if port > 65535 {
			return 0
		}
		port += 1
	}
}

func isPortAvailable(port int) bool {
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return true
	} else {
		defer conn.Close()
		return false
	}
}

func writePortToFile(port int) error {
	exeDir, err := getExecutableDir()
	if err != nil {
		return err
	}
	portFile := path.Join(exeDir, "port")
	return ioutil.WriteFile(portFile, []byte(strconv.Itoa(port)+"\n"), 0644)

}

func getExecutableDir() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	return path.Dir(ex), nil
}
