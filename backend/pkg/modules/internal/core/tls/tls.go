package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/openclarity/apiclarity/backend/pkg/config"
)

type ClientTLSOptions struct {
	CustomTLSTransport *http.Transport
}

func CreateClientTLSOptions(conf *config.Config) (*ClientTLSOptions, error) {
	if !conf.EnableTLS {
		return nil, nil
	}

	tlsTransport, err := createTLSTransport(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS transport: %v", err)
	}

	return &ClientTLSOptions{
		CustomTLSTransport: tlsTransport,
	}, nil
}

func createTLSTransport(conf *config.Config) (*http.Transport, error) {
	// Get the SystemCertPool.
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadFile(conf.RootCertFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%v): %v", conf.RootCertFilePath, err)
	}

	// Append provided root cert to the system pool.
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("failed to append certs from PEM")
	}

	// Trust the augmented cert pool in our client.
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = tlsConfig

	return customTransport, nil
}
