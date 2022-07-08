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

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

func ResyncedModule(wrappedModuleFactory core.ModuleFactory) core.ModuleFactory {
	return func(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
		wrappedModule, err := wrappedModuleFactory(ctx, accessor)
		if err != nil {
			return nil, err
		}
		rec := NewRecovery(wrappedModule.Info().Name, accessor)
		m := &module{
			Module:   wrappedModule,
			eventsCh: make(chan *core.Event),
		}
		eventsCh := rec.Resync(ctx, m.eventsCh)
		go func() {
			for {
				select {
				case <-ctx.Done():
					log.Error(ctx.Err())
					return
				case event, ok := <-eventsCh:
					if !ok {
						return
					}
					go wrappedModule.EventNotify(ctx, event)
				}
			}
		}()
		return m, nil
	}
}

type module struct {
	core.Module
	eventsCh chan *core.Event
}

func (m module) EventNotify(ctx context.Context, event *core.Event) { m.eventsCh <- event }
