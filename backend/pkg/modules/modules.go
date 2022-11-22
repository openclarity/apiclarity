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
	"fmt"

	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/apiclarity/backend/pkg/backend/speculatoraccessor"
	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/database"

	// Enables the bfla module.
	_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"

	// Enables the fuzzer module.
	_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer"

	// Enables the spec differ module.
	_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/spec_differ"

	// Enables the spec reconstructor module.
	_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/specreconstructor"

	// Enables the traceanalyzer module.
	_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer"
	"github.com/openclarity/apiclarity/backend/pkg/notifier"
	"github.com/openclarity/apiclarity/backend/pkg/sampling"
)

type (
	ModuleInfo          = core.ModuleInfo
	ModulesManager      = core.Module //nolint:revive
	MockModule          = core.MockModule
	Annotation          = core.Annotation
	BackendAccessor     = core.BackendAccessor
	MockBackendAccessor = core.MockBackendAccessor
	Event               = core.Event
)

var (
	NewMockModulesManager  = core.NewMockModule
	NewMockBackendAccessor = core.NewMockBackendAccessor
)

func New(ctx context.Context, dbHandler *database.Handler, clientset kubernetes.Interface, samplingManager *sampling.TraceSamplingManager, speculatorAccessor speculatoraccessor.SpeculatorsAccessor, notifier *notifier.Notifier, config *config.Config) (ModulesManager, []ModuleInfo, error) {
	backendAccessor, err := core.NewAccessor(dbHandler, clientset, samplingManager, speculatorAccessor, notifier, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create backend accessor: %v", err)
	}

	module, infos := core.New(ctx, backendAccessor, samplingManager, config.TraceSamplingEnabled)

	return module, infos, nil
}
