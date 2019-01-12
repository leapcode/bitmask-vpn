package go_locale

import (
    "os/exec"
)

func getCommandOutput(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)

    out, err := cmd.Output()
    if err != nil {
        return "", err
    }

    return string(out), nil
}
