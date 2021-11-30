package helper

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
		"--tls-version-min", "1.2",
		"--float",
		"--log", filepath.Join(LogFolder, "openvpn-leap.log"),
	}

	allowedArgs = map[string][]string{
		"--remote":            []string{"IP", "NUMBER", "PROTO"},
		"--tls-cipher":        []string{"CIPHER"},
		"--cipher":            []string{"CIPHER"},
		"--auth":              []string{"CIPHER"},
		"--management-client": []string{},
		"--management":        []string{"IP", "NUMBER"},
		"--route":             []string{"IP", "IP", "NETGW"},
		"--cert":              []string{"FILE"},
		"--key":               []string{"FILE"},
		"--ca":                []string{"FILE"},
		"--fragment":          []string{"NUMBER"},
		"--keepalive":         []string{"NUMBER", "NUMBER"},
		"--verb":              []string{"NUMBER"},
		"--tun-ipv6":          []string{},
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
			log.Printf("Invalid openvpn arg: %s", args[i])
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
			log.Printf("Invalid openvpn arg params: %v", args[i:i+len(params)+1])
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
