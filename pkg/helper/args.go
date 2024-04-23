package helper

import (
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	nameserverTCP = "10.41.0.1"
	nameserverUDP = "10.42.0.1"
)

var (
	fixedArgs = []string{
		"--nobind",
		"--client",
		"--dev", "tun",
		"--tls-client",
		"--remote-cert-tls", "server",
		"--dhcp-option", "DNS", nameserverTCP,
		"--dhcp-option", "DNS", nameserverUDP,
		"--tls-version-min", "1.0",
		"--float",
		"--log", filepath.Join(LogFolder, "openvpn-leap.log"),
	}

	allowedArgs = map[string][]string{
		"--remote":            {"IP", "NUMBER", "PROTO"},
		"--tls-cipher":        {"CIPHER"},
		"--cipher":            {"CIPHER"},
		"--auth":              {"CIPHER"},
		"--management-client": {},
		"--management":        {"IP", "NUMBER"},
		"--route":             {"IP", "IP", "NETGW"},
		"--cert":              {"FILE"},
		"--key":               {"FILE"},
		"--ca":                {"FILE"},
		"--fragment":          {"NUMBER"},
		"--keepalive":         {"NUMBER", "NUMBER"},
		"--verb":              {"NUMBER"},
		"--tun-ipv6":          {},
	}

	cipher  = regexp.MustCompile("^[A-Z0-9-]+$")
	formats = map[string]func(s string) bool{
		"NUMBER": isNumber,
		"PROTO":  isProto,
		"IP":     isIP,
		"CIPHER": cipher.MatchString,
		"FILE":   isFile,
		"NETGW":  isNetGw,
	}
)

func parseOpenvpnArgs(args []string) []string {
	newArgs := fixedArgs
	newArgs = append(newArgs, getPlatformOpenvpnFlags()...)
	for i := 0; i < len(args); i++ {
		params, ok := allowedArgs[args[i]]
		if !ok {
			log.Warn().
				Str("arg", args[i]).
				Msg("Invalid openvpn arg")
			continue
		}
		for j, arg := range args[i+1 : i+len(params)+1] {
			if !formats[params[j]](arg) {
				ok = false
				break
			}
		}
		if ok {
			newArgs = append(newArgs, args[i:i+len(params)+1]...)
			i = i + len(params)
		} else {
			log.Warn().
				Str("params", strings.Join(args[i:i+len(params)+1], " ")).
				Msg("Invalid openvpn arg params")
		}
	}
	return newArgs
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func isProto(s string) bool {
	for _, proto := range []string{"tcp", "udp", "tcp4", "udp4", "tcp6", "udp6"} {
		if s == proto {
			return true
		}
	}
	return false
}

func isIP(s string) bool {
	return net.ParseIP(s) != nil
}

func isFile(s string) bool {
	info, err := os.Stat(s)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isNetGw(s string) bool {
	return s == "net_gateway"
}
