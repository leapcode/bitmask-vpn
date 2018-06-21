package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"0xacab.org/leap/bitmask-systray/bitmask"
)

var pidFile = filepath.Join(bitmask.ConfigPath, "systray.pid")

func acquirePID() error {
	pid := syscall.Getpid()
	current, err := getPID()
	if err != nil {
		return err
	}

	if current != pid && pidRunning(current) {
		return fmt.Errorf("Another systray is running with pid: %d", current)
	}

	return setPID(pid)
}

func releasePID() error {
	pid := syscall.Getpid()
	current, err := getPID()
	if err != nil {
		return err
	}
	if current != 0 && current != pid {
		return fmt.Errorf("Can't release pid file, is not own by this process")
	}

	if current == pid {
		return os.Remove(pidFile)
	}
	return nil
}

func getPID() (int, error) {
	_, err := os.Stat(pidFile)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	file, err := os.Open(pidFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return strconv.Atoi(string(b))
}

func setPID(pid int) error {
	file, err := os.Create(pidFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	return err
}

func pidRunning(pid int) bool {
	if pid == 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if runtime.GOOS == "windows" {
		return err == nil
	}

	if err != nil {
		log.Printf("An error ocurred finding process: %v", err)
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true
	}
	return false
}
