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
	"strings"

	"github.com/openclarity/apiclarity/backend/pkg/config"
	"github.com/openclarity/trace-sampling-manager/manager/pkg/manager"
	log "github.com/sirupsen/logrus"
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

func GetNotificationPrefix() string {
	notificationPrefix, ok := os.LookupEnv(config.NotificationPrefix)
	if !ok {
		return ""
	}
	return notificationPrefix
}

// The order of the modules is not important.
// You MUST NOT rely on a specific order of modules.
var modules map[string]ModuleFactory = map[string]ModuleFactory{}

func RegisterModule(m ModuleFactory) {
	_, corePath, _, _ := runtime.Caller(0)
	_, modulePath, _, _ := runtime.Caller(1)
	modulePathIndex := len(strings.Split(corePath, "/")) - 2
	moduleID := strings.Split(modulePath, "/")[modulePathIndex]

	modules[moduleID] = m
}

type ModuleFactory func(ctx context.Context, moduleName string, accessor BackendAccessor) (Module, error)

func New(ctx context.Context, accessor BackendAccessor, samplingManager *manager.Manager) (Module, []ModuleInfo) {
	c := &Core{}
	c.Modules = map[string]Module{}
	c.samplingManager = samplingManager

	modInfos := []ModuleInfo{}
	for moduleName, moduleFactory := range modules {
		module, err := moduleFactory(ctx, moduleName, accessor)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Module %s initialized", module.Info().Name)
		c.Modules[moduleName] = module
		modInfos = append(modInfos, module.Info())
	}
	return c, modInfos
}

type Core struct {
	Modules         map[string]Module
	samplingManager *manager.Manager
}

func (c *Core) Info() ModuleInfo {
	return ModuleInfo{
		Name:        "Core",
		Description: "Core Module Stub",
	}
}

func shouldTrace(host string, hosts []string) bool {
	for _, h := range hosts {
		if h == "*" || h == host {
			return true
		}
	}
	return false
}

func (c *Core) EventNotify(ctx context.Context, event *Event) {

	host := event.APIEvent.HostSpecName
	for modName, mod := range c.Modules {

		if !shouldTrace(host, c.samplingManager.HostsToTraceByComponentID(modName)) {
			log.Debugf("Trace of host %s should NOT be sent to module %s.", host, modName)
			continue
		}
		log.Debugf("Trace of host %s should be sent to module %s.", host, modName)

		mod.EventNotify(ctx, event)
	}
}

func (c *Core) HTTPHandler() http.Handler {
	handler := http.NewServeMux()
	for moduleName, m := range c.Modules {
		handler.Handle(BaseHTTPPath+"/"+moduleName+"/", m.HTTPHandler())
	}

	return handler
}
