package obfsvpn

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	CERT     = "cert"
	IAT_MODE = "iat-mode"
	// some implementations use iatMode instead of iat-mode /o\
	IAT_MODE2 = "iatMode"
)

var OBFS4_REGEX = compileObfs4Regex()

// AddStringParameter returns a regex pattern string based on the provided key.
func addStringParameter(key string) string {
	return fmt.Sprintf("(?:\\s+(%s=\\S+))?", key)
}

// AddIntParameter returns a regex pattern string based on the provided key.
func addIntParameter(key string) string {
	return fmt.Sprintf("(?:(\\s+%s=\\d+))?", key)
}

func compileObfs4Regex() *regexp.Regexp {
	pattern := fmt.Sprintf("^((obfs4\\s+(\\S+)\\s+(\\S+))|(Bridge obfs4 <IP ADDRESS>:<PORT> <FINGERPRINT>))(%s|%s|%s)*\\s*$",
		addStringParameter(CERT),
		addIntParameter(IAT_MODE),
		addIntParameter(IAT_MODE2),
	)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}
	return regex
}

// ParseObfs4CertFromBridgelineFile reads the specified file and extracts the obfs4 certificate.
func ParseObfs4CertFromBridgelineFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		cert := GetCertFromLine(line)
		if len(cert) > 0 {
			return cert, nil
		}
	}

	return "", fmt.Errorf("certificate not found in the file")
}

func GetCertFromLine(line string) string {
	matches := OBFS4_REGEX.FindStringSubmatch(line)
	if len(matches) > 0 {
		var certValue string
		for _, submatch := range matches {
			if len(submatch) > 5 && submatch[:5] == "cert=" {
				certValue = submatch[5:] // Get the value after "cert="
				return strings.TrimSuffix(certValue, "==")
			}
		}
	}
	return ""
}
