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

package modules

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/apiclarity/apiclarity/backend/pkg/database"

	// Enables the bfla module.
	_ "github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"

	// Enables the demo module.
	_ "github.com/apiclarity/apiclarity/backend/pkg/modules/internal/demo"
)

type (
	Module              = core.Module
	MockModule          = core.MockModule
	Annotation          = core.Annotation
	BackendAccessor     = core.BackendAccessor
	MockBackendAccessor = core.MockBackendAccessor
	Event               = core.Event
)

var (
	NewMockModule          = core.NewMockModule
	NewMockBackendAccessor = core.NewMockBackendAccessor
)

func New(ctx context.Context, dbHandler *database.Handler, clientset kubernetes.Interface) Module {
	return core.New(ctx, core.NewAccessor(dbHandler, clientset))
}
