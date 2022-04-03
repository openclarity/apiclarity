package backend

import (
	"github.com/apiclarity/apiclarity/api/server/models"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/golang/mock/gomock"
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
	NewProvidedSpec string
	OldProvidedSpec string

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
	//if event.IsNonAPI != m.IsNonAPI {
	//	return false
	//}
	return true
}
