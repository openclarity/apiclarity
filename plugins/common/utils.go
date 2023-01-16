// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package common

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	tracesamplingclient "github.com/openclarity/trace-sampling-manager/api/client/client"

	"github.com/openclarity/apiclarity/plugins/api/client/client"
	"github.com/openclarity/apiclarity/plugins/api/client/models"
)

const (
	MaxBodySize              = 1000 * 1000
	RequestIDHeaderKey       = "X-Request-Id"
	RequestTimeContextKey    = "request_time"
	SamplingInterval         = 10 * time.Second
	MinimumSeparatedHostSize = 2
)

func ReadBody(body io.ReadCloser) ([]byte, bool, error) {
	ret, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read body: %v", err)
	}
	if len(ret) > MaxBodySize {
		return []byte{}, true, nil
	}
	return ret, false, nil
}

func GetRequestIDFromHeadersOrGenerate(reqHeaders http.Header) string {
	if reqID, ok := reqHeaders[RequestIDHeaderKey]; ok {
		return reqID[0]
	}
	// no request id header, generate request id
	requestID := uuid.NewV4()

	return requestID.String()
}

func CreateHeaders(headers map[string][]string) []*models.Header {
	ret := []*models.Header{}

	// TODO support multiple values for a header
	for header, values := range headers {
		ret = append(ret, &models.Header{
			Key:   header,
			Value: values[0],
		})
	}
	return ret
}

func GetPathWithQuery(reqURL *url.URL) string {
	pathAndQuery := reqURL.Path
	if !strings.HasPrefix(pathAndQuery, "/") {
		pathAndQuery = "/" + pathAndQuery
	}
	if reqURL.RawQuery != "" {
		pathAndQuery += "?" + reqURL.RawQuery
	}
	return pathAndQuery
}

const CACertFile = "/etc/traces/certs/root-ca.crt"

type ClientTLSOptions struct {
	RootCAFileName string
}

func NewTelemetryAPIClient(host string, tlsOptions *ClientTLSOptions) (*client.APIClarityPluginsTelemetriesAPI, error) {
	var clientTransport runtime.ClientTransport
	var err error

	if tlsOptions != nil {
		clientTransport, err = createClientTransportTLS(host, tlsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create tls client transport: %v", err)
		}
	} else {
		clientTransport = httptransport.New(host, client.DefaultBasePath, []string{"http"})
	}

	apiClient := client.New(clientTransport, strfmt.Default)
	return apiClient, nil
}

func NewTraceSamplingAPIClient(host string, tlsOptions *ClientTLSOptions) (*tracesamplingclient.TraceSamplingManager, error) {
	var clientTransport runtime.ClientTransport
	var err error

	if tlsOptions != nil {
		clientTransport, err = createClientTransportTLS(host, tlsOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to create tls client transport: %v", err)
		}
	} else {
		clientTransport = httptransport.New(host, client.DefaultBasePath, []string{"http"})
	}

	apiClient := tracesamplingclient.New(clientTransport, strfmt.Default)
	return apiClient, nil
}

func createClientTransportTLS(host string, tlsOptions *ClientTLSOptions) (runtime.ClientTransport, error) {
	// Get the SystemCertPool
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert file
	certs, err := ioutil.ReadFile(tlsOptions.RootCAFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file (%v): %v", tlsOptions.RootCAFileName, err)
	}

	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("failed to append certs from PEM")
	}

	//Trust the augmented cert pool in our client
	tlsConfig := &tls.Config{
		RootCAs:            rootCAs,
		InsecureSkipVerify: true,
	}
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = tlsConfig

	transport := httptransport.NewWithClient(host, client.DefaultBasePath, []string{"https"},
		&http.Client{Transport: customTransport})

	return transport, nil
}

// If host from URL is not containing any dots, the defaultNamespace will be added to the host name
func GetHostAndPortFromURL(URL, defaultNamespace string) (host, port string) {
	if !strings.Contains(URL, "://") {
		// need to add scheme to host in order for url.Parse to parse properly
		URL = "http://" + URL
	}

	parsedHost, err := url.Parse(URL)
	if err != nil {
		log.Errorf("Failed to parse URL=%v: %v", URL, err)
		return URL, ""
	}

	host = parsedHost.Hostname()
	port = parsedHost.Port()

	host = strings.TrimSuffix(host, ".svc.cluster.local")
	host = strings.TrimSuffix(host, ".svc.cluster")
	host = strings.TrimSuffix(host, ".svc")

	// we assume that host with no dots is an internal host, so we will "fix" it with default namespace.
	if !strings.Contains(host, ".") && defaultNamespace != "" {
		host = host + "." + defaultNamespace
	}

	if port == "" {
		if parsedHost.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	return
}

// Will try to extract the namespace from the host name, and if not found, will use the provided default namespace.
func GetDestinationNamespaceFromHostOrDefault(host, defaultNamespace string) string {
	if sp := strings.Split(host, "."); len(sp) >= MinimumSeparatedHostSize {
		return sp[1]
	}
	return defaultNamespace
}

func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
