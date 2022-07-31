// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = tlsConfig

	return customTransport, nil
}
