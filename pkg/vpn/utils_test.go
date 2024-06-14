package vpn

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
}

func TestIsValidCertExpired(t *testing.T) {
	certFile := "testdata/expired.pem"
	require.False(t, isValidCert(certFile), "The test with the expired pem failed")
}

func TestIsValidCertEmpty(t *testing.T) {
	certFile := "testdata/empty.pem"
	require.False(t, isValidCert(certFile), "The test with the empty pem file failed")
}

func TestIsValidCertKeyMissing(t *testing.T) {
	certFile := "testdata/privatekeymissing.pem"
	require.False(t, isValidCert(certFile), "The test with the missing private key failed")
}

func TestIsValidCertBroken(t *testing.T) {
	certFile := "testdata/broken.pem"
	require.False(t, isValidCert(certFile), "The test with the broken pem file failed")
}

