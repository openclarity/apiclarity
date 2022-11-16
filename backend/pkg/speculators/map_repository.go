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

package speculators

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

type Repository struct {
	Speculators      map[uint]*_speculator.Speculator
	speculatorConfig _speculator.Config

	lock *sync.RWMutex
}

func NewMapRepository(config _speculator.Config) *Repository {
	return &Repository{
		Speculators:      map[uint]*_speculator.Speculator{},
		speculatorConfig: config,
		lock:             &sync.RWMutex{},
	}
}

func DecodeState(filePath string, config _speculator.Config) (*Repository, error) {
	r := Repository{}

	const perm = 400
	file, err := os.OpenFile(filePath, os.O_RDONLY, os.FileMode(perm))
	if err != nil {
		return nil, fmt.Errorf("failed to open file (%v): %v", filePath, err)
	}
	defer closeFile(file)

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode state: %v", err)
	}

	r.speculatorConfig = config
	r.lock = &sync.RWMutex{}
	log.Info("Speculator state was decoded")
	// log.Debugf("Speculator Config %+v", config)

	return &r, nil
}

func (r *Repository) Get(speculatorID uint) *_speculator.Speculator {
	r.lock.RLock()
	defer r.lock.RUnlock()

	speculator, ok := r.Speculators[speculatorID]
	if !ok {
		r.Speculators[speculatorID] = _speculator.CreateSpeculator(r.speculatorConfig)

		return r.Speculators[speculatorID]
	}

	return speculator
}

func (r *Repository) EncodeState(filePath string) error {
	const perm = 400
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.FileMode(perm))
	if err != nil {
		return fmt.Errorf("failed to open state file: %v", err)
	}
	defer closeFile(file)

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(r)
	if err != nil {
		return fmt.Errorf("failed to encode state: %v", err)
	}

	return nil
}

func closeFile(file *os.File) {
	if err := file.Close(); err != nil {
		log.Errorf("Failed to close file: %v", err)
	}
}
