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

package recovery

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/database"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	ack               = "ack"
	maxEventsPerQuery = 100
	retryDelay        = 1 * time.Second
)

// nolint:revive
func NewRecovery(modName string, accessor core.BackendAccessor) *recovery {
	return &recovery{
		accessor:   accessor,
		modName:    modName,
		limit:      maxEventsPerQuery,
		retryDelay: retryDelay,
	}
}

type recovery struct {
	retryDelay time.Duration
	accessor   core.BackendAccessor
	modName    string
	limit      int
}

func (r *recovery) Resync(ctx context.Context, events chan *core.Event) (allEvents chan *core.Event) {
	pastEvents := r.getUnackedEvents(ctx)
	allEvents = make(chan *core.Event)

	go func() {
		var liveEvents []*core.Event
		persistedEvents := map[uint]*core.Event{}

	loop:
		for {
			select {
			case event := <-events:
				liveEvents = append(liveEvents, event)
			case event, ok := <-pastEvents:
				if !ok {
					// switch to live translation
					break loop
				}

				persistedEvents[event.ID] = &core.Event{APIEvent: event}

			case <-ctx.Done():
				log.Error(ctx.Err())
				return
			}
		}

		log.Infof("synced %d past events", len(persistedEvents))

		// catch up
		for _, e := range persistedEvents {
			allEvents <- e
		}
		for _, e := range liveEvents {
			// if the event was registered and sent do not send it again
			if _, ok := persistedEvents[e.APIEvent.ID]; ok {
				continue
			}
			allEvents <- e
		}

		log.Debugf("start live transmission")
		for {
			select {
			case event, ok := <-events:
				if !ok {
					return
				}
				log.Infof("push event: %d", event.APIEvent.ID)
				allEvents <- event
			case <-ctx.Done():
				log.Error(ctx.Err())
				return
			}
		}
	}()
	return allEvents
}

func (r *recovery) getAPIEvents(ctx context.Context, filter database.GetAPIEventsQuery) []*database.APIEvent {
	for {
		events, err := r.accessor.GetAPIEvents(ctx, filter)
		if err != nil {
			log.Errorf("error getting api events: %s; retrying after %d", err, r.retryDelay)
			time.Sleep(r.retryDelay)
			continue
		}
		return events
	}
}

func (r *recovery) getUnackedEvents(ctx context.Context) chan *database.APIEvent {
	unackedEvents := make(chan *database.APIEvent)
	ack := ack
	go func() {
		offset := 0
		for {
			events := r.getAPIEvents(ctx, database.GetAPIEventsQuery{
				Offset: offset,
				Limit:  r.limit,
				APIEventAnnotationFilters: &database.APIEventAnnotationFilters{
					NoAnnotations: true,
					ModuleNameIs:  &r.modName,
					NameIsNot:     &ack,
				},
			})
			for _, event := range events {
				unackedEvents <- event
			}
			eventsCount := len(events)
			if eventsCount < r.limit || eventsCount == 0 {
				close(unackedEvents)
				return
			}
			offset += eventsCount
			log.Infof("got %d events", offset)
		}
	}()
	return unackedEvents
}
