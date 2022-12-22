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
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

var (
	TestSpanStartTime      = time.Date(2022, 2, 11, 20, 26, 12, 321, time.UTC)
	TestSpanStartTimestamp = pcommon.NewTimestampFromTime(TestSpanStartTime)

	TestSpanEventTime      = time.Date(2022, 2, 11, 20, 26, 13, 123, time.UTC)
	TestSpanEventTimestamp = pcommon.NewTimestampFromTime(TestSpanEventTime)

	TestSpanEndTime      = time.Date(2022, 2, 11, 20, 26, 13, 789, time.UTC)
	TestSpanEndTimestamp = pcommon.NewTimestampFromTime(TestSpanEndTime)
)

var (
	resourceAttributes1  = map[string]interface{}{"resource-attr": "resource-attr-val-1"}
	spanEventAttributes  = map[string]interface{}{"span-event-attr": "span-event-attr-val"}
	spanClientAttributes = []map[string]interface{}{
		{
			"span.kind":        "client",
			"http.method":      "PUT",
			"http.flavor":      "1.1",
			"http.url":         "https://api.thecorporation.com:8443/treasure/chest/100",
			"net.peer.ip":      "192.0.2.5",
			"http.status_code": 200,
		},
	}
	spanServerAttributes = []map[string]interface{}{
		{
			// minimum attributes
			"span.kind":      "server",
			"http.method":    "GET",
			"http.url":       "/catalogue/808a2de1-1aaa-4c25-a9b9-6612e8f29a38",
			"http.route":     "/catalogue/:id",
			"http.client_ip": "34.23.65.32",
		},
		{
			// complete attributes
			"span.kind":        "server",
			"http.scheme":      "http",
			"http.method":      "GET",
			"http.target":      "/treasure/chest/100",
			"http.host":        "api.thecorporation.com",
			"http.route":       "/treasure/chest/:treasure_id",
			"http.server_name": "sr234.eu-west.aws.com",
			"http.client_ip":   "34.23.65.32",
			"net.host.port":    80,
			"net.peer.ip":      "192.0.25.4",
			"http.status_code": 200,
		},
	}
)

// Function from "github.com/open-telemetry/opentelemetry-collector-contrib/internal/common/testutil"
func GetAvailableLocalAddress(t testing.TB, network string) string {
	ln, err := net.Listen(network, "localhost:0")
	require.NoError(t, err, "Failed to get a free local port")
	// There is a possible race if something else takes this same port before
	// the test uses it, however, that is unlikely in practice.
	defer func() {
		assert.NoError(t, ln.Close())
	}()
	return ln.Addr().String()
}

// Function from "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata/trace"
func fillSpanOne(span ptrace.Span, kind ptrace.SpanKind) {
	span.SetName("operationA")
	span.SetStartTimestamp(TestSpanStartTimestamp)
	span.SetEndTimestamp(TestSpanEndTimestamp)
	span.SetDroppedAttributesCount(1)
	span.SetKind(kind)
	evs := span.Events()
	ev0 := evs.AppendEmpty()
	ev0.SetTimestamp(TestSpanEventTimestamp)
	ev0.SetName("event-with-attr")
	eventAttrs := pcommon.NewMap()
	eventAttrs.FromRaw(spanEventAttributes)
	eventAttrs.CopyTo(ev0.Attributes())
	ev0.SetDroppedAttributesCount(2)
	ev1 := evs.AppendEmpty()
	ev1.SetTimestamp(TestSpanEventTimestamp)
	ev1.SetName("event")
	ev1.SetDroppedAttributesCount(2)
	span.SetDroppedEventsCount(1)
	status := span.Status()
	status.SetCode(ptrace.StatusCodeOk)
	status.SetMessage("status-ok")
}

func fillSpanAttributes(span ptrace.Span, spanAttributes map[string]interface{}) {
	attrs := span.Attributes()
	for key, value := range spanAttributes {
		switch value := value.(type) {
		case string:
			attrs.PutString(key, value)
		case int:
			attrs.PutInt(key, int64(value))
		}
	}
}

func generateTracesOneSpan() ptrace.Traces {
	td := ptrace.NewTraces()
	td.ResourceSpans().AppendEmpty()
	rs0 := td.ResourceSpans().At(0)
	resAttrs := pcommon.NewMap()
	resAttrs.FromRaw(resourceAttributes1)
	resAttrs.CopyTo(rs0.Resource().Attributes())
	td.ResourceSpans().At(0).ScopeSpans().AppendEmpty()
	rs0ils0 := td.ResourceSpans().At(0).ScopeSpans().At(0)
	rs0ils0.Spans().AppendEmpty()
	return td
}

// Function from "github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
func GenerateTracesOneSpan(kind ptrace.SpanKind) ptrace.Traces {
	td := generateTracesOneSpan()
	rs0ils0 := td.ResourceSpans().At(0).ScopeSpans().At(0)
	span := rs0ils0.Spans().At(0)
	fillSpanOne(span, kind)
	return td
}

func TracesOneSpanAddAttributes(td ptrace.Traces, attrs map[string]interface{}) ptrace.Traces {
	rs0ils0 := td.ResourceSpans().At(0).ScopeSpans().At(0)
	span := rs0ils0.Spans().At(0)
	fillSpanAttributes(span, attrs)
	return td
}
