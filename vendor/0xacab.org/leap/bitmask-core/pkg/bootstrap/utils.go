package bootstrap

import (
	"crypto/sha256"
	gotls "crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	bitmask_storage "0xacab.org/leap/bitmask-core/pkg/storage"
	"0xacab.org/leap/menshen/models"
	"github.com/go-openapi/runtime"
	openapi "github.com/go-openapi/runtime/client"
	utls "github.com/refraction-networking/utls"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
)

// Parses API URL of menshen. Returns hostname/ip, port, useTLS
func parseApiURL(menshenURL string) (string, int, bool, error) {
	url, err := url.Parse(menshenURL)
	if err != nil {
		return "", -1, false, fmt.Errorf("Could not parse API url %s: %s", url, err)
	}

	hostname := url.Hostname()
	useTLS := url.Scheme != "http"

	var port int
	if url.Port() == "" {
		port = 443
	} else {
		port, err = strconv.Atoi(url.Port())
		if err != nil {
			return "", -1, false, fmt.Errorf("Could not parse port to int %s: %s", url.Port(), err)
		}
	}

	log.Trace().
		Bool("useTLS", useTLS).
		Str("hostname", hostname).
		Int("port", port).
		Msg("Parsed API URL")

	return hostname, port, useTLS, nil
}

func (c *Config) getAPIClient() *http.Client {
	if c.UseTLS {
		client := &http.Client{
			Transport: &http2.Transport{
				// Hook into TLS connection buildup to resolve IP with DNS over HTTP (DoH)
				DialTLS: func(network, addr string, tlsCfg *gotls.Config) (net.Conn, error) {
					if c.ResolveWithDoH {
						log.Debug().
							Str("domain", addr).
							Msg("Resolving host with DNS over HTTPs")

						ip, err := dohQuery(c.Host)
						if err != nil {
							return nil, err
						}

						log.Debug().
							Str("domain", addr).
							Str("ip", ip).
							Msg("Sucessfully resolved host via DNS over HTTPs")
						if strings.Contains(ip, ":") {
							// IPv6 address requires extra brackets in order to
							// distinguish address from port
							addr = fmt.Sprintf("[%s]:%d", ip, c.Port)
						} else {
							addr = fmt.Sprintf("%s:%d", ip, c.Port)
						}
					}

					roller, err := utls.NewRoller()
					if err != nil {
						return nil, err
					}
					uconn, err := roller.Dial(network, addr, c.Host)
					if err != nil {
						return nil, err
					}

					uconn.SetSNI(c.Host)
					return uconn, err
				},
			},
			Timeout: time.Duration(30) * time.Second,
		}
		return client
	} else {
		return &http.Client{Timeout: time.Duration(30) * time.Second}
	}
}

// Returns authentication header (invite token) from database
// Returns nil if no introducer is saved or an error occurs
func (api *API) getInviteTokenAuth() runtime.ClientAuthInfoWriter {
	if len(api.config.Introducer) == 0 {
		return nil
	}

	log.Trace().Msg("Getting invite token from db")
	storage, err := bitmask_storage.GetStorage()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Could not get storage to load invite token")
		return nil
	}

	introducer, err := storage.GetIntroducerByFQDN(api.config.Host)
	if err != nil {
		log.Debug().
			Str("err", err.Error()).
			Str("fqdn", api.config.Host).
			Msg("Could not get introducer by fqdn")
		return nil
	}

	if len(introducer.Auth) == 0 {
		log.Warn().Msg("An introducer was found for this fqdn, but the invite token is empty")
		return nil
	}

	log.Debug().Msg("Sending invite token")
	return openapi.APIKeyAuth("x-menshen-auth-token", "header", introducer.Auth)
}

func SupportsApiv5(provider *models.ModelsProvider) bool {
	for _, version := range provider.APIVersions {
		if version == "5" {
			return true
		}
	}
	return false
}

func ShouldUpdateEIPService(unixTimestamp int64) bool {
	timestamp := time.Unix(unixTimestamp, 0)
	return time.Since(timestamp) > 72*time.Hour
}

func ShouldUpdateBridges(unixTimestamp int64) bool {
	timestamp := time.Unix(unixTimestamp, 0)
	return time.Since(timestamp) > 4*time.Hour
}

func ShouldUpdateGateways(unixTimestamp int64) bool {
	timestamp := time.Unix(unixTimestamp, 0)
	return time.Since(timestamp) > 24*time.Hour
}

func ShouldUpdateOpenVPNCredentials(credentials, caCertFingerprint string) bool {
	caCrt := GetCAFromApi5OpenvpnCertResponse(credentials)
	ovpnCrt := GetCertFromApi5OpenvpnCertResponse(credentials)
	key := GetKeyFromApi5OpenvpnCertResponse(credentials)

	if caCrt == "" || ovpnCrt == "" || key == "" {
		log.Debug().Msg("OpenVPN credentials are missing.")
		return true
	}

	tomorrow := time.Now().In(time.UTC).Add(24 * time.Hour)
	if err := ValidateCertificate(ovpnCrt, tomorrow); err != nil {
		log.Warn().Msgf("OpenVPN certificate is not valid: %v", err)
		return true
	}

	if err := ValidateCertificateWithFP(caCrt, caCertFingerprint, tomorrow); err != nil {
		log.Warn().Msgf("CA certificate is not valid: %v", err)
		return true
	}

	return true
}

// Validate a certificate against an expiry date
func ValidateCertificate(pemCert string, minimalTimeBoreExpiry time.Time) error {
	return ValidateCertificateWithFP(pemCert, "", minimalTimeBoreExpiry)
}

// Validate a certificate against a SHA256 fingerprint and an expiry date
func ValidateCertificateWithFP(pemCert string, fingerprint string, minimalTimeBoreExpiry time.Time) error {
	crtBlock, _ := pem.Decode([]byte(pemCert))
	if crtBlock == nil {
		return fmt.Errorf("could not decode pem certificate")
	}

	cert, err := x509.ParseCertificate(crtBlock.Bytes)
	if err != nil {
		return fmt.Errorf("could not parse certificate")
	}

	expires := cert.NotAfter

	if !expires.After(minimalTimeBoreExpiry) {
		return fmt.Errorf("certificate is expired: %v expected minimalTimeBeforeExpiry: %v", expires, minimalTimeBoreExpiry)
	}

	if fingerprint != "" {
		// normalize fingerprint
		fingerprint = strings.ToLower(fingerprint)
		fingerprint = strings.TrimPrefix(fingerprint, "sha256:")
		fingerprint = strings.TrimSpace(fingerprint)

		digest := sha256.Sum256(crtBlock.Bytes)
		hexString := strings.TrimSpace(hex.EncodeToString(digest[:]))
		if hexString != fingerprint {
			return fmt.Errorf("cert fingerprint does not match: %v expected: %v", hexString, fingerprint)
		}
	}
	return nil
}

func ToJson(model any) (string, error) {
	res, err := json.Marshal(model)
	if err != nil {
		return "", err
	}
	return string(res), nil
}
