package go_locale

import (
    "strings"
    "strconv"
)

func DetectLocale() (string, error) {
    out, err := getCommandOutput("wmic", "os", "get", "locale")
    if err != nil {
        return "", err
    }

    out = strings.Replace(out, "Locale", "", -1)
    out = strings.TrimSpace(out)

    id, err := strconv.ParseInt(out, 16, 64)
    if err != nil {
        return "", err
    }

    lcid := LCID()
    locale, err := lcid.ById(int(id))
    if err != nil {
        return "", err
    }

    return locale, nil
}
