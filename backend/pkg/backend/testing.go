package backend

import (
	"github.com/apiclarity/apiclarity/api/server/models"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/golang/mock/gomock"
)

const (
	specKey            = "httpbin:8080"
	host               = "httpbin"
	port               = "8080"
	destinationAddress = "1.1.1.1:8080"
)

type eventMatcher struct {
	Method                   models.HTTPMethod
	Path                     string
	ProvidedPathID           string
	ReconstructedPathID      string
	Query                    string
	StatusCode               int64
	SourceIP                 string
	DestinationIP            string
	DestinationPort          int64
	HasReconstructedSpecDiff bool
	HasProvidedSpecDiff      bool
	HasSpecDiff              bool
	SpecDiffType             models.DiffType
	HostSpecName             string
	IsNonAPI                 bool

	NewReconstructedSpec string
	OldReconstructedSpec string
	NewProvidedSpec      string
	OldProvidedSpec      string

	APIInfoID uint
	EventType models.APIType
}

func NewEventMatcher(event _database.APIEvent) gomock.Matcher {
	return &eventMatcher{
		Method:                   event.Method,
		Path:                     event.Path,
		ProvidedPathID:           event.ProvidedPathID,
		ReconstructedPathID:      event.ReconstructedPathID,
		Query:                    event.Query,
		StatusCode:               event.StatusCode,
		SourceIP:                 event.SourceIP,
		DestinationIP:            event.DestinationIP,
		DestinationPort:          event.DestinationPort,
		HasReconstructedSpecDiff: event.HasReconstructedSpecDiff,
		HasProvidedSpecDiff:      event.HasProvidedSpecDiff,
		HasSpecDiff:              event.HasSpecDiff,
		SpecDiffType:             event.SpecDiffType,
		HostSpecName:             event.HostSpecName,
		IsNonAPI:                 event.IsNonAPI,
		NewReconstructedSpec:     event.NewReconstructedSpec,
		OldReconstructedSpec:     event.OldReconstructedSpec,
		NewProvidedSpec:          event.NewProvidedSpec,
		OldProvidedSpec:          event.OldProvidedSpec,
		APIInfoID:                event.APIInfoID,
		EventType:                event.EventType,
	}
}

func (m *eventMatcher) String() string {
	return "Event Matcher"
}

func (m *eventMatcher) Matches(x interface{}) bool {
	event, ok := x.(*_database.APIEvent)
	if !ok {
		return false
	}
	if event.Method != m.Method {
		return false
	}
	if event.Path != m.Path {
		return false
	}
	if event.HostSpecName != m.HostSpecName {
		return false
	}
	if event.Query != m.Query {
		return false
	}
	if event.StatusCode != m.StatusCode {
		return false
	}
	if event.SourceIP != m.SourceIP {
		return false
	}
	if event.DestinationIP != m.DestinationIP {
		return false
	}
	if event.DestinationPort != m.DestinationPort {
		return false
	}
	if event.IsNonAPI != m.IsNonAPI {
		return false
	}
	if event.EventType != m.EventType {
		return false
	}
	if event.SpecDiffType != m.SpecDiffType {
		return false
	}
	if event.HasProvidedSpecDiff != m.HasProvidedSpecDiff {
		return false
	}
	return true
}

type APIEventTest struct {
	event _database.APIEvent
}

func createDefaultTestEvent() *APIEventTest {
	return &APIEventTest{
		event: _database.APIEvent{
			Method:                   "GET",
			Path:                     "/test",
			ReconstructedPathID:      "",
			Query:                    "foo=bar",
			StatusCode:               200,
			SourceIP:                 "2.2.2.2",
			DestinationIP:            "1.1.1.1",
			DestinationPort:          8080,
			HasReconstructedSpecDiff: false,
			HasProvidedSpecDiff:      false,
			HasSpecDiff:              false,
			SpecDiffType:             models.DiffTypeNODIFF,
			HostSpecName:             host,
			IsNonAPI:                 false,
			NewReconstructedSpec:     "",
			OldReconstructedSpec:     "",
			NewProvidedSpec:          "",
			OldProvidedSpec:          "",
			APIInfoID:                0,
			EventType:                models.APITypeINTERNAL,
		},
	}
}

func (t *APIEventTest) WithEventType(eventType models.APIType) *APIEventTest {
	t.event.EventType = eventType
	return t
}

func (t *APIEventTest) WithSpecDiffType(diffType models.DiffType) *APIEventTest {
	t.event.SpecDiffType = diffType
	return t
}

func (t *APIEventTest) WithIsNonAPI(isNonAPI bool) *APIEventTest {
	t.event.IsNonAPI = isNonAPI
	return t
}

func (t *APIEventTest) WithHasProvidedSpecDiff(hasProvidedSpecDiff bool) *APIEventTest {
	t.event.HasProvidedSpecDiff = hasProvidedSpecDiff
	return t
}

func (t *APIEventTest) WithHasReconstructedSpecDiff(hasReconstructedSpecDiff bool) *APIEventTest {
	t.event.HasReconstructedSpecDiff = hasReconstructedSpecDiff
	return t
}

