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

package core

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/apiclarity/backend/pkg/sampling"
)

const BaseHTTPPath = "/api/modules"

// GetAssetsDir get assets directory from env variable or the default location.
func GetAssetsDir() string {
	assetsDir, ok := os.LookupEnv(config.ModulesAssetsEnvVar)
	if !ok {
		_, file, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(file), "..", "..", "assets")
	}
	return assetsDir
}

// The order of the modules is not important.
// You MUST NOT rely on a specific order of modules.
var modules = map[string]ModuleFactory{}

func RegisterModule(m ModuleFactory) {
	_, corePath, _, ok1 := runtime.Caller(0)
	_, modulePath, _, ok2 := runtime.Caller(1)
	if !ok1 || !ok2 {
		log.Errorf("unable to retrieve folder containing the module %v. Ignoring registration.", m)
		return
	}
	modulePathIndex := len(strings.Split(corePath, "/")) - 2 //nolint:gomnd
	moduleFolderName := strings.Split(modulePath, "/")[modulePathIndex]

	modules[moduleFolderName] = m
}

type ModuleFactory func(ctx context.Context, accessor BackendAccessor) (Module, error)

func New(ctx context.Context, accessor BackendAccessor, samplingManager *sampling.TraceSamplingManager, traceSamplingEnabled bool) (Module, []ModuleInfo) {
	c := &Core{}
	c.Modules = map[string]Module{}
	c.samplingManager = samplingManager
	c.traceSamplingEnabled = traceSamplingEnabled

	modInfos := []ModuleInfo{}
	for moduleFolderName, moduleFactory := range modules {
		module, err := moduleFactory(ctx, accessor)
		if err != nil {
			log.Error(err)
			continue
		}
		if module.Info().Name != moduleFolderName {
			log.Panicf("module %s's name does not match with folder name containing the module %s", module.Info().Name, moduleFolderName)
		}
		log.Infof("Module %s initialized", module.Info().Name)
		c.Modules[moduleFolderName] = module
		modInfos = append(modInfos, module.Info())
	}
	return c, modInfos
}

type Core struct {
	Modules              map[string]Module
	samplingManager      *sampling.TraceSamplingManager
	traceSamplingEnabled bool
}

func (c *Core) Info() ModuleInfo {
	return ModuleInfo{
		Name:        "Core",
		Description: "Core Module Stub",
	}
}

func shouldTrace(host string, port int64, hosts []string) bool {
	for _, h := range hosts {
		if h == "*" {
			return true
		}
		hp := strings.Split(h, ":")
		if hp[0] == "*" || hp[0] == host {
			if len(hp) == 1 {
				return true
			}
			if hp[1] == "*" {
				return true
			}
			p, err := strconv.Atoi(hp[1])
			if err == nil && int64(p) == port {
				return true
			}
		}
	}
	return false
}

func (c *Core) EventNotify(ctx context.Context, event *Event) {
	host := event.APIEvent.HostSpecName
	port := event.APIEvent.DestinationPort
	traceSourceID := event.APIInfo.TraceSourceID

	for modName, mod := range c.Modules {
		if c.traceSamplingEnabled {
			hosts, err := c.samplingManager.GetHostsToTrace(modName, traceSourceID)
			if err != nil {
				log.Debugf("Failed to retrieve hosts for traceSource %v for module %s.", traceSourceID, modName)
				continue
			}
			if !shouldTrace(host, port, hosts) {
				log.Debugf("Trace of host %s should NOT be sent to module %s.", host, modName)
				continue
			}
		}
		log.Debugf("Trace of host %s should be sent to module %s.", host, modName)

		mod.EventNotify(ctx, event)
	}
}

func (c *Core) HTTPHandler() http.Handler {
	handler := http.NewServeMux()
	for moduleName, m := range c.Modules {
		if m.HTTPHandler() != nil {
			handler.Handle(BaseHTTPPath+"/"+moduleName+"/", m.HTTPHandler())
		}
	}

	return handler
}
