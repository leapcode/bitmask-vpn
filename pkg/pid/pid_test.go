package pid

import (
	"syscall"
	"testing"
)

const (
	invalidPid = 345678
)

func TestPidRunning(t *testing.T) {
	pid := syscall.Getpid()
	if !pidRunning(pid) {
		t.Errorf("pid %v is not running", pid)
	}

	if pidRunning(invalidPid) {
		t.Errorf("pid %v is running", invalidPid)
	}
}
