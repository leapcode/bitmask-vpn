package go_locale

import (
    "strings"
    "errors"
)

func DetectLocale() (string, error) {
    out, err := getCommandOutput("locale")
    if err != nil {
        return "", err
    }

    lines := strings.Split(out, "\n")
    for _, line := range lines {
        if line != "" {
            parts := strings.Split(line, "=")
            value := strings.Trim(parts[1], `"`)

            if value != "C" && value != "" {
                encodingIndex := strings.Index(value, ".")
                if encodingIndex != -1 {
                    value = value[0:encodingIndex]
                }
                return value, nil
            }
        }
    }

    return "", errors.New("unable to locale locale")
}
