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

package apiclarityexporter

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	apiclientmodels "github.com/openclarity/apiclarity/plugins/api/client/models"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

const (
	missingAttrValue     string = "<missing>"
	DefaultSourceAddress string = "client:5280"
	DefaultStatusCode    string = "200"
)

func wrapAttributeError(logger *zap.Logger, msg, attrKey, attrValue string, err error) error {
	logger.Debug(msg,
		zap.String("attribute", attrKey),
		zap.String(attrKey, attrValue),
		zap.Error(err),
	)
	return fmt.Errorf("%s, attribute: %s, value: %s, error: %w", msg, attrKey, attrValue, err)
}

func parseResourceServerAttrs(actel *apiclientmodels.Telemetry, resource pcommon.Resource) bool {
	ok := true
	resAttrs := resource.Attributes()
	if ipAddr, ok := resAttrs.Get("ip"); ok {
		actel.DestinationAddress = ipAddr.AsString()
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if serviceIP, ok := resAttrs.Get(string("ipv4")); ok {
		actel.DestinationAddress = serviceIP.AsString()
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if hostName, ok := resAttrs.Get(string(semconv.HostNameKey)); ok {
		actel.DestinationAddress = hostName.AsString()
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if serviceName, ok := resAttrs.Get(string(semconv.ServiceNameKey)); ok {
		actel.DestinationAddress = serviceName.AsString()
	} else {
		ok = false
	}
	return ok
}

func setTelemetryClientSpan(actel *apiclientmodels.Telemetry, resource pcommon.Resource, attrs pcommon.Map, logger *zap.Logger) error {
	//Set destination/server address
	if peerName, ok := attrs.Get(string(semconv.NetPeerNameKey)); ok {
		actel.DestinationAddress = peerName.AsString()
		if portAttr, portOk := attrs.Get(string(semconv.NetPeerPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if peerIP, ok := attrs.Get(string(semconv.NetPeerIPKey)); ok {
		actel.DestinationAddress = peerIP.AsString()
		if portAttr, portOk := attrs.Get(string(semconv.NetPeerPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if actel.Request.Host != "" {
		//Assume this is from URL or Host header...
		actel.DestinationAddress = actel.Request.Host
	} else if ok := parseResourceServerAttrs(actel, resource); !ok {
		//Either HTTPURLKey, HTTPHostKey, NetPeerNameKey or NetPeerIPKey should be defined
		return wrapAttributeError(logger, "missing attribute", string(semconv.NetPeerIPKey), missingAttrValue, nil)
	}

	//Set source/client address
	if hostIpAttr, ok := attrs.Get(string(semconv.NetHostIPKey)); ok {
		actel.SourceAddress = hostIpAttr.AsString()
	} else if hostNameAttr, ok := attrs.Get(string(semconv.NetHostNameKey)); ok {
		actel.SourceAddress = hostNameAttr.AsString()
	}
	if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
		actel.SourceAddress = actel.SourceAddress + ":" + portAttr.AsString()
	}

	return nil
}

func setTelemetryServerSpan(actel *apiclientmodels.Telemetry, resource pcommon.Resource, attrs pcommon.Map, logger *zap.Logger) error {
	//Set destination/server address
	if serverNameAttr, ok := attrs.Get(string(semconv.HTTPServerNameKey)); ok {
		actel.DestinationAddress = serverNameAttr.AsString()
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if hostNameAttr, ok := attrs.Get(string(semconv.NetHostNameKey)); ok {
		actel.DestinationAddress = hostNameAttr.AsString()
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if hostIPAttr, ok := attrs.Get(string(semconv.NetHostIPKey)); ok {
		actel.DestinationAddress = hostIPAttr.AsString()
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if actel.Request.Host != "" {
		//Assume this is from URL or Host header...
		actel.DestinationAddress = actel.Request.Host
	} else if ok := parseResourceServerAttrs(actel, resource); !ok {
		//Either HTTPURLKey, HTTPHostKey, HTTPServerNameKey or NetHostNameKey should be defined
		return wrapAttributeError(logger, "missing attribute", string(semconv.HTTPServerNameKey), missingAttrValue, nil)
	}

	//Set source/client address
	if clientIP, ok := attrs.Get(string(semconv.HTTPClientIPKey)); ok {
		actel.SourceAddress = clientIP.AsString()
	} else if peerName, ok := attrs.Get(string(semconv.NetPeerNameKey)); ok {
		actel.SourceAddress = peerName.AsString()
	} else if peerIP, ok := attrs.Get(string(semconv.NetPeerIPKey)); ok {
		actel.SourceAddress = peerIP.AsString() // this could be a proxy
	}
	if portAttr, portOk := attrs.Get(string(semconv.NetPeerPortKey)); portOk {
		actel.SourceAddress = actel.SourceAddress + ":" + portAttr.AsString()
	}

	return nil
}

// Process a single span into APIClarity telemetry
func (e *exporter) processOTelSpan(resource pcommon.Resource, _ pcommon.InstrumentationScope, span ptrace.Span) (*apiclientmodels.Telemetry, error) {
	/*
		res.Attributes().Range(func(k string, v pcommon.Value) bool {
			e.logger.Debug("Checking resource attributes",
				zap.String("key", k),
				zap.String("value", v.AsString()),
			)
			return true
		})
	*/
	e.logger.Info("Converting span",
		zap.String("kind", span.Kind().String()),
		zap.String("name", span.Name()),
		zap.String("traceid", span.TraceID().HexString()),
		zap.Int("attributes.length", span.Attributes().Len()),
	)

	span.Attributes().Range(func(k string, v pcommon.Value) bool {
		e.logger.Debug("Checking span attributes",
			zap.String("key", k),
			zap.String("value", v.AsString()),
		)
		return true
	})

	req := &apiclientmodels.Request{
		Common: &apiclientmodels.Common{
			TruncatedBody: false,
			Time:          span.StartTimestamp().AsTime().Unix(),
			Headers:       []*apiclientmodels.Header{},
		},
	}
	resp := &apiclientmodels.Response{
		Common: &apiclientmodels.Common{
			TruncatedBody: false,
			Time:          span.EndTimestamp().AsTime().Unix(),
			Headers:       []*apiclientmodels.Header{},
		},
	}
	actel := &apiclientmodels.Telemetry{
		DestinationAddress: "",
		SourceAddress:      "",
		Request:            req,
		Response:           resp,
	}

	attrs := span.Attributes()

	var urlOk bool
	var urlAttr pcommon.Value
	if urlAttr, urlOk = attrs.Get(string(semconv.HTTPURLKey)); urlOk {
		urlVal := urlAttr.StringVal()
		if urlVal == "" {
			urlOk = false
		} else {
			urlInfo, err := url.Parse(urlVal)
			if err != nil {
				return nil, wrapAttributeError(e.logger, "cannot parse attribute", string(semconv.HTTPURLKey), urlVal, err)
			}
			actel.Scheme = urlInfo.Scheme
			actel.Request.Host = urlInfo.Host
			actel.Request.Path = urlInfo.Path
		}
	}
	if schemeAttr, schemeOk := attrs.Get(string(semconv.HTTPSchemeKey)); schemeOk {
		actel.Scheme = schemeAttr.AsString()
	} else if !urlOk {
		//Either HTTPURLKey or HTTPSchemeKey should be defined
		return nil, wrapAttributeError(e.logger, "missing attribute", string(semconv.HTTPSchemeKey), missingAttrValue, nil)
	}
	if targetAttr, targetOk := attrs.Get(string(semconv.HTTPTargetKey)); targetOk {
		actel.Request.Path = targetAttr.AsString()
	} else if !urlOk {
		//Either HTTPURLKey or HTTPTargetKey should be defined
		return nil, wrapAttributeError(e.logger, "missing attribute", string(semconv.HTTPTargetKey), missingAttrValue, nil)
	}
	//Do not override URL with Host header, but check for use later
	if hostAttr, hostOk := attrs.Get(string(semconv.HTTPHostKey)); hostOk && actel.Request.Host == "" {
		actel.Request.Host = hostAttr.AsString() // host is Host Header. Is this correct?
	}

	var err error
	switch span.Kind() {
	case ptrace.SpanKindClient:
		err = setTelemetryClientSpan(actel, resource, attrs, e.logger)
	case ptrace.SpanKindServer:
		err = setTelemetryServerSpan(actel, resource, attrs, e.logger)
	default:
		e.logger.Debug("ignoring span because it is not client or server",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
		)
	}
	if err != nil {
		span.Attributes().Range(func(k string, v pcommon.Value) bool {
			e.logger.Warn("Failing span attribute",
				zap.String("key", k),
				zap.String("value", v.AsString()),
			)
			return true
		})
		return nil, err
	}

	//Speculator requires address to have a port?
	if !strings.Contains(actel.DestinationAddress, ":") {
		if actel.Scheme == "http" {
			actel.DestinationAddress = actel.DestinationAddress + ":80"
		} else if actel.Scheme == "https" {
			actel.DestinationAddress = actel.DestinationAddress + ":443"
		} else {
			e.logger.Warn("Cannot infer destination port, using default 80",
				zap.String("kind", span.Kind().String()),
				zap.String("name", span.Name()),
				zap.String("traceid", span.TraceID().HexString()),
			)
			actel.DestinationAddress = actel.DestinationAddress + ":80"
		}
	}
	if actel.SourceAddress == "" {
		e.logger.Warn("Cannot infer source address, using default",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
			zap.String("address", DefaultSourceAddress),
		)
		actel.SourceAddress = DefaultSourceAddress
	} else if !strings.Contains(actel.SourceAddress, ":") {
		_, defaultPort, _ := strings.Cut(DefaultSourceAddress, ":")
		e.logger.Warn("Cannot infer source port, using default",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
			zap.String("port", defaultPort),
		)
		actel.SourceAddress = actel.SourceAddress + ":" + defaultPort
	}
	//APIClarity requires a host?
	if actel.Request.Host == "" {
		e.logger.Warn("Cannot find host, using destination",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
			zap.String("destination", actel.DestinationAddress),
		)
		actel.Request.Host = actel.DestinationAddress
	}

	// Fill in missing data where available.
	if method, ok := attrs.Get(string(semconv.HTTPMethodKey)); ok {
		actel.Request.Method = method.AsString()
	}
	if statusCode, ok := attrs.Get(string(semconv.HTTPStatusCodeKey)); ok {
		actel.Response.StatusCode = statusCode.AsString()
	} else {
		e.logger.Warn("Cannot find status code, using default",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
			zap.String(string(semconv.HTTPStatusCodeKey), DefaultStatusCode),
		)
		actel.Response.StatusCode = DefaultStatusCode
	}
	if flavor, ok := attrs.Get(string(semconv.HTTPFlavorKey)); ok {
		actel.Request.Common.Version = flavor.AsString()
		actel.Response.Common.Version = flavor.AsString()
	}
	if route, ok := attrs.Get(string(semconv.HTTPRouteKey)); ok {
		actel.Request.Path = route.AsString()
	}

	attrs.Range(func(k string, v pcommon.Value) bool {
		e.logger.Debug("Converting span attributes",
			zap.String("key", k),
			zap.String("value", v.AsString()),
		)
		// Convert header formats
		s := strings.TrimPrefix(k, "http.request.header.")
		if len(s) < len(k) {
			actel.Request.Common.Headers = append(actel.Request.Common.Headers, &apiclientmodels.Header{
				Key:   strings.ReplaceAll(s, "_", "-"),
				Value: v.AsString(),
			})
			return true
		}
		s = strings.TrimPrefix(k, "http.response.header.")
		if len(s) < len(k) {
			actel.Response.Common.Headers = append(actel.Response.Common.Headers, &apiclientmodels.Header{
				Key:   strings.ReplaceAll(s, "_", "-"),
				Value: v.AsString(),
			})
			return true
		}
		return true
	})

	// After parsing headers, we could check if the request id is already there...
	idGen, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot create request id for telemetry: %w", err)
	}
	actel.RequestID = idGen.String()

	return actel, nil
}
