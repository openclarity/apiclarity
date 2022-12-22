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
	"net"
	"net/url"
	"strings"

	"github.com/gofrs/uuid"
	apiclientmodels "github.com/openclarity/apiclarity/plugins/api/client/models"
	"go.uber.org/zap"

	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.9.0"
)

const (
	missingAttrValue     = "<missing>"
	DefaultSourceAddress = "client:5280"
	DefaultStatusCode    = "200"
	DefaultScheme        = "http"
	DefaultTarget        = "/"
	DefaultSpanKind      = ptrace.SpanKindServer
	RequestBody          = attribute.Key("request_body")
	ResponseBody         = attribute.Key("response_body")
)

func wrapAttributeError(logger *zap.Logger, msg, attrKey, attrValue string, err error) error {
	logger.Debug(msg,
		zap.String("attribute", attrKey),
		zap.String(attrKey, attrValue),
		zap.Error(err),
	)
	return fmt.Errorf("%s, attribute: %s, value: %s, error: %w", msg, attrKey, attrValue, err)
}

func (e *exporter) convertAddr(addr string) string {
	//TODO: make it configurable to prefer IP or hostname
	isIpAddr := net.ParseIP(addr) != nil
	if isIpAddr && e.config.PreferHostNames {
		if aliases, err := net.LookupAddr(addr); err != nil && len(aliases) > 0 {
			e.logger.Info("lookup IP to get hostname",
				zap.String("address", addr),
				zap.String("host", aliases[0]),
			)
			return aliases[0]
		} else if err != nil {
			e.logger.Info("failed lookup IP to get hostname",
				zap.String("address", addr),
				zap.Error(err),
			)
		}
	} else if !isIpAddr && !e.config.PreferHostNames {
		if hosts, err := net.LookupHost(addr); err != nil && len(hosts) > 0 {
			e.logger.Info("lookup hostname to get IP",
				zap.String("address", addr),
				zap.String("host", hosts[0]),
			)
			return hosts[0]
		} else if err != nil {
			e.logger.Info("failed lookup hostname to get IP",
				zap.String("address", addr),
				zap.Error(err),
			)
		}
	} else {
		e.logger.Info("address is already in preferred form",
			zap.String("address", addr),
		)
	}
	return addr
}

func (e *exporter) convertHost(addr string) string {
	if addr == "" {
		return addr
	}
	//Manage prefs for IP v. hostname for <ip|host>[:port]
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		e.logger.Error("failed to split host/port",
			zap.String("address", addr),
			zap.Error(err),
		)
		return addr
	}
	if host != "" {
		host = e.convertAddr(host)
	}
	if host != "" && port != "" {
		return host + ":" + port
	} else if host != "" {
		return host
	} else {
		return ":" + port
	}
}

func (e *exporter) parseResourceServerAttrs(actel *apiclientmodels.Telemetry, resource pcommon.Resource) bool {
	ok := true
	resAttrs := resource.Attributes()
	if ipAddr, ok := resAttrs.Get("ip"); ok {
		actel.DestinationAddress = e.convertAddr(ipAddr.AsString())
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if serviceIP, ok := resAttrs.Get(string("ipv4")); ok {
		actel.DestinationAddress = e.convertAddr(serviceIP.AsString())
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if hostName, ok := resAttrs.Get(string(semconv.HostNameKey)); ok {
		actel.DestinationAddress = e.convertAddr(hostName.AsString())
		if servicePort, ok := resAttrs.Get("port"); ok {
			actel.DestinationAddress = actel.DestinationAddress + ":" + servicePort.AsString()
		}
	} else if serviceName, ok := resAttrs.Get(string(semconv.ServiceNameKey)); ok {
		actel.DestinationAddress = serviceName.AsString()
	} else {
		ok = false //nolint:ineffassign
	}
	return ok
}

func (e *exporter) setTelemetryClientSpan(actel *apiclientmodels.Telemetry, resource pcommon.Resource, attrs pcommon.Map, logger *zap.Logger) error {
	//Set destination/server address
	if peerName, ok := attrs.Get(string(semconv.NetPeerNameKey)); ok {
		actel.DestinationAddress = e.convertAddr(peerName.AsString())
		if portAttr, portOk := attrs.Get(string(semconv.NetPeerPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if peerIP, ok := attrs.Get(string(semconv.NetPeerIPKey)); ok {
		actel.DestinationAddress = e.convertAddr(peerIP.AsString())
		if portAttr, portOk := attrs.Get(string(semconv.NetPeerPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if actel.Request.Host != "" {
		//Assume this is from URL or Host header...
		//TODO: split addr/port and convertAddr
		actel.DestinationAddress = actel.Request.Host
	} else if ok := e.parseResourceServerAttrs(actel, resource); !ok {
		//Either HTTPURLKey, HTTPHostKey, NetPeerNameKey or NetPeerIPKey should be defined
		return wrapAttributeError(logger, "missing attribute", string(semconv.NetPeerIPKey), missingAttrValue, nil)
	}

	//Set source/client address
	if hostIpAttr, ok := attrs.Get(string(semconv.NetHostIPKey)); ok {
		actel.SourceAddress = e.convertAddr(hostIpAttr.AsString())
	} else if hostNameAttr, ok := attrs.Get(string(semconv.NetHostNameKey)); ok {
		actel.SourceAddress = e.convertAddr(hostNameAttr.AsString())
	}
	if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
		actel.SourceAddress = actel.SourceAddress + ":" + portAttr.AsString()
	}

	return nil
}

func (e *exporter) setTelemetryServerSpan(actel *apiclientmodels.Telemetry, resource pcommon.Resource, attrs pcommon.Map, logger *zap.Logger) error {
	//Set destination/server address
	if serverNameAttr, ok := attrs.Get(string(semconv.HTTPServerNameKey)); ok {
		actel.DestinationAddress = e.convertAddr(serverNameAttr.AsString())
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if hostNameAttr, ok := attrs.Get(string(semconv.NetHostNameKey)); ok {
		actel.DestinationAddress = e.convertAddr(hostNameAttr.AsString())
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if hostIPAttr, ok := attrs.Get(string(semconv.NetHostIPKey)); ok {
		actel.DestinationAddress = e.convertAddr(hostIPAttr.AsString())
		if portAttr, portOk := attrs.Get(string(semconv.NetHostPortKey)); portOk {
			actel.DestinationAddress = actel.DestinationAddress + ":" + portAttr.AsString()
		}
	} else if actel.Request.Host != "" {
		//Assume this is from URL or Host header...
		actel.DestinationAddress = actel.Request.Host
	} else if ok := e.parseResourceServerAttrs(actel, resource); !ok {
		//Either HTTPURLKey, HTTPHostKey, HTTPServerNameKey or NetHostNameKey should be defined
		return wrapAttributeError(logger, "missing attribute", string(semconv.HTTPServerNameKey), missingAttrValue, nil)
	}

	//Set source/client address
	if clientIP, ok := attrs.Get(string(semconv.HTTPClientIPKey)); ok {
		actel.SourceAddress = e.convertAddr(clientIP.AsString())
	} else if peerName, ok := attrs.Get(string(semconv.NetPeerNameKey)); ok {
		actel.SourceAddress = e.convertAddr(peerName.AsString())
	} else if peerIP, ok := attrs.Get(string(semconv.NetPeerIPKey)); ok {
		actel.SourceAddress = e.convertAddr(peerIP.AsString()) //this could be a proxy
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
		urlVal := urlAttr.Str()
		if urlVal == "" {
			urlOk = false
		} else {
			urlInfo, err := url.Parse(urlVal)
			if err != nil {
				return nil, wrapAttributeError(e.logger, "cannot parse attribute", string(semconv.HTTPURLKey), urlVal, err)
			}
			actel.Scheme = urlInfo.Scheme
			actel.Request.Host = e.convertHost(urlInfo.Host)
			actel.Request.Path = urlInfo.Path
		}
	}
	if schemeAttr, schemeOk := attrs.Get(string(semconv.HTTPSchemeKey)); schemeOk {
		actel.Scheme = schemeAttr.AsString()
	} else if !urlOk {
		//Either HTTPURLKey or HTTPSchemeKey should be defined
		actel.Scheme = DefaultScheme
		e.logger.Warn("Using default value for missing attribute",
			zap.String("attribute", string(semconv.HTTPSchemeKey)),
			zap.String("value", DefaultScheme),
		)
	}
	//Some frameworks use http.path although it's not in the semconv
	if path, ok := attrs.Get("http.path"); ok {
		actel.Request.Path = path.AsString()
	} else if targetAttr, targetOk := attrs.Get(string(semconv.HTTPTargetKey)); targetOk {
		actel.Request.Path = targetAttr.AsString()
	} else if !urlOk {
		//Either HTTPURLKey or HTTPTargetKey should be defined
		actel.Request.Path = DefaultTarget
		e.logger.Warn("Using default value for missing attribute",
			zap.String("attribute", string(semconv.HTTPTargetKey)),
			zap.String("value", DefaultTarget),
		)
	}
	//Do not override URL with Host header, but check for use later
	if hostAttr, hostOk := attrs.Get(string(semconv.HTTPHostKey)); hostOk && actel.Request.Host == "" {
		actel.Request.Host = e.convertHost(hostAttr.AsString()) // host is Host Header. Is this correct?
	}

	var err error
	switch span.Kind() {
	case ptrace.SpanKindClient:
		err = e.setTelemetryClientSpan(actel, resource, attrs, e.logger)
	case ptrace.SpanKindServer:
		err = e.setTelemetryServerSpan(actel, resource, attrs, e.logger)
	case ptrace.SpanKindUnspecified:
		e.logger.Warn("span kind unspecified, assuming default",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
			zap.Int("default", int(DefaultSpanKind)),
		)
		if DefaultSpanKind == ptrace.SpanKindClient {
			err = e.setTelemetryClientSpan(actel, resource, attrs, e.logger)
		} else {
			err = e.setTelemetryServerSpan(actel, resource, attrs, e.logger)
		}
	default:
		e.logger.Warn("ignoring span that is not client or server",
			zap.String("kind", span.Kind().String()),
			zap.String("name", span.Name()),
			zap.String("traceid", span.TraceID().HexString()),
		)
	}
	if err != nil {
		span.Attributes().Range(func(k string, v pcommon.Value) bool {
			e.logger.Warn("failing span attribute",
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
		parts := strings.Split(DefaultSourceAddress, ":")
		defaultPort := parts[1]
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
		//actel.Request.ParameterizedPath = route.AsString()
	}

	// Add payloads if available
	if reqBody, ok := attrs.Get(string(RequestBody)); ok {
		if reqBody.Type() == pcommon.ValueTypeBytes {
			actel.Request.Common.Body = reqBody.Bytes().AsRaw()
		} else if reqBody.Type() == pcommon.ValueTypeStr {
			actel.Request.Common.Body = []byte(reqBody.Str())
		} else {
			e.logger.Warn("unknown request body value type",
				zap.String("kind", span.Kind().String()),
				zap.String("name", span.Name()),
				zap.String("traceid", span.TraceID().HexString()),
				zap.String("type", reqBody.Type().String()),
			)
		}
		//TODO: check media type
		if actel.Request.Common.Body != nil {
			actel.Request.Common.Headers = append(actel.Request.Common.Headers, &apiclientmodels.Header{
				Key:   "Content-Type",
				Value: "application/json",
			})
		}
	}
	if respBody, ok := attrs.Get(string(ResponseBody)); ok {
		if respBody.Type() == pcommon.ValueTypeBytes {
			actel.Response.Common.Body = respBody.Bytes().AsRaw()
		} else if respBody.Type() == pcommon.ValueTypeStr {
			actel.Response.Common.Body = []byte(respBody.Str())
		} else {
			e.logger.Warn("unknown response body value type",
				zap.String("kind", span.Kind().String()),
				zap.String("name", span.Name()),
				zap.String("traceid", span.TraceID().HexString()),
				zap.String("type", respBody.Type().String()),
			)
		}
		//TODO: check media type
		if actel.Response.Common.Body != nil {
			actel.Response.Common.Headers = append(actel.Response.Common.Headers, &apiclientmodels.Header{
				Key:   "Content-Type",
				Value: "application/json",
			})
		}
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
