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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

type SetState func(state interface{})

//go:generate go run github.com/golang/mock/mockgen -package=recovery -destination=./mocks.gen.go . StatePersister

type StatePersister interface {
	UseState(apiID uint, name string, val interface{}) (setFn SetState, found bool, err error)
	Keys(name string) []uint
	Persist(ctx context.Context) error
	AckSubmit(eventID uint)
}

func NewStatePersister(ctx context.Context, accessor core.BackendAccessor, modName string, persistInterval time.Duration) StatePersister {
	p := &persister{
		statesMu: &sync.RWMutex{},
		states:   map[uint]map[string]*stateValue{},
		eventsMu: &sync.RWMutex{},
		accessor: accessor,
		modName:  modName,
	}
	go func() {
		ticker := time.NewTicker(persistInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Error(ctx.Err())
				return
			case <-ticker.C:
				if err := p.persistAPIInfoAnnotations(ctx); err != nil {
					log.Error(err)
				}

				if err := p.persistAPIEventAnnotations(ctx); err != nil {
					log.Error(err)
				}
			}
		}
	}()
	return p
}

type persister struct {
	modName  string
	statesMu *sync.RWMutex
	states   map[uint]map[string]*stateValue
	eventsMu *sync.RWMutex
	events   []uint
	accessor core.BackendAccessor
}

type stateValue struct {
	stateChanged bool
	val          interface{}
}

type Errors []error

func (errs Errors) Error() string {
	buff := bytes.NewBufferString("errors: \n")
	for _, err := range errs {
		buff.WriteString(err.Error())
		buff.WriteString("\n")
	}
	return buff.String()
}

func (p *persister) Persist(ctx context.Context) error {
	if err := p.persistAPIInfoAnnotations(ctx); err != nil {
		return err
	}

	return p.persistAPIEventAnnotations(ctx)
}

func (p *persister) persistAPIEventAnnotations(ctx context.Context) error {
	p.eventsMu.Lock()
	defer p.eventsMu.Unlock()
	var errs Errors
	acked := 0
	events := p.events
	for ; len(p.events) > 0; p.events = p.events[:len(p.events)-1] {
		eventID := p.events[len(p.events)-1]
		if err := p.accessor.CreateAPIEventAnnotations(ctx, p.modName, eventID, core.Annotation{
			Name: ack, Annotation: []byte("true"),
		}); err != nil {
			errs = append(errs, err)
			continue
		}
		acked++
	}
	log.Debugf("Acked %d events; events: %d", acked, events)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (p *persister) persistAPIInfoAnnotations(ctx context.Context) error {
	p.statesMu.Lock()
	defer p.statesMu.Unlock()

	var errs Errors

	annsPersisted := 0
	for key, values := range p.states {
		var anns []core.Annotation
		for name, val := range values {
			// ignore annotations that didn't change state
			if !val.stateChanged {
				continue
			}
			valBytes, err := json.Marshal(val.val)
			if err != nil {
				errs = append(errs, err)
				continue
			}

			anns = append(anns, core.Annotation{
				Name:       name,
				Annotation: valBytes,
			})
		}
		if len(anns) == 0 {
			log.Debugf("nothing to update for module=%s; apiID=%d", p.modName, key)
			continue
		}
		log.Debugf("store api info moduleName=%s apiID=%d", p.modName, key)
		err := p.accessor.StoreAPIInfoAnnotations(ctx, p.modName, key, anns...)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// signal that the state has been persisted
		for _, ann := range anns {
			p.states[key][ann.Name].stateChanged = false
		}
		annsPersisted += len(anns)
	}
	log.Debugf("Persisted %d annotations", annsPersisted)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (p *persister) Keys(name string) (keys []uint) {
	p.statesMu.RLock()
	defer p.statesMu.RUnlock()
	for key, annNameAndValue := range p.states {
		for annName := range annNameAndValue {
			if name == annName {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

func (p *persister) UseState(apiID uint, name string, val interface{}) (setFn SetState, found bool, err error) {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, false, errors.New("val is not a pointer or is nil")
	}
	setFn = func(state interface{}) {
		log.Debugf("set new state moduleName=%s apiID=%d", p.modName, apiID)
		// set the new state
		p.statesMu.Lock()
		p.states[apiID][name] = &stateValue{stateChanged: true, val: state}
		p.statesMu.Unlock()
	}
	p.statesMu.Lock()
	defer p.statesMu.Unlock()
	_, ok := p.states[apiID]
	if !ok {
		p.states[apiID] = map[string]*stateValue{}
	}
	_, ok = p.states[apiID][name]
	if !ok {
		ann, err := p.accessor.GetAPIInfoAnnotation(context.TODO(), p.modName, apiID, name)
		if err != nil {
			return setFn, false, fmt.Errorf("unable to get api info annotations for: modName=%s apiID=%d name=%s: %w", p.modName, apiID, name, err)
		}
		if ann != nil {
			if err := json.Unmarshal(ann.Annotation, val); err != nil {
				return setFn, false, fmt.Errorf("unable to unmarshal json: %w", err)
			}
			p.states[apiID][name] = &stateValue{val: rv.Elem().Interface()}
		}
		return setFn, true, nil
	}
	st, ok := p.states[apiID][name]
	rv.Elem().Set(reflect.ValueOf(st.val))
	return setFn, ok, nil
}

func (p *persister) AckSubmit(eventID uint) {
	log.Debugf("ack submit: %d", eventID)
	p.eventsMu.Lock()
	p.events = append(p.events, eventID)
	p.eventsMu.Unlock()
}

type T = interface{}

type PersistedMap interface {
	Get(apiID uint) (PersistedValue, error)
	Keys() []uint
}

type PersistedValue interface {
	Exists() bool
	Get() T
	Set(val T)
}

func NewPersistedMap(sp StatePersister, name string, valType reflect.Type) PersistedMap {
	return &persistedMap{name: name, sp: sp, valType: valType}
}

type persistedMap struct {
	valType reflect.Type
	sp      StatePersister
	name    string
}

func (p *persistedMap) Keys() []uint {
	return p.sp.Keys(p.name)
}

func (p *persistedMap) Get(apiID uint) (PersistedValue, error) {
	val := reflect.New(p.valType)
	set, found, err := p.sp.UseState(apiID, p.name, val.Interface())
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to get state: %w", err)
		}
		found = false
	}

	pv := &persistedValue{set: set, found: found, valType: p.valType}
	if found {
		pv.val = val.Interface()
	}

	return pv, nil
}

type persistedValue struct {
	found   bool
	val     T // cache the value
	set     SetState
	valType reflect.Type
}

func (p *persistedValue) Exists() bool { return p.found }
func (p *persistedValue) Get() (t T) {
	if p.val != nil {
		return reflect.ValueOf(p.val).Elem().Interface()
	}
	return reflect.Zero(p.valType).Interface()
}

func (p *persistedValue) Set(val T) {
	p.set(val)
	p.val = val
}
