package core

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/apiclarity/apiclarity/backend/pkg/config"
)

const BaseHTTPPath = "/api/modules"

//GetAssetsDir get assets directory from env variable or the default location
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
var modules []ModuleFactory

func RegisterModule(m ModuleFactory) {
	modules = append(modules, m)
}

type ModuleFactory func(ctx context.Context, accessor BackendAccessor) (Module, error)

func New(ctx context.Context, accessor BackendAccessor) *core {
	c := &core{}
	for _, moduleFactory := range modules {
		module, err := moduleFactory(ctx, accessor)
		if err != nil {
			log.Error(err)
			continue
		}
		log.Infof("Module %s initialized", module.Name())
		c.modules = append(c.modules, module)
	}
	return c
}

type core struct {
	modules []Module
}

func (c *core) Name() string { return "core" }

func (c *core) EventNotify(ctx context.Context, event *Event) {
	for _, mod := range c.modules {
		mod.EventNotify(ctx, event)
	}
}

func (c *core) HTTPHandler() http.Handler {
	handler := http.NewServeMux()
	for _, m := range c.modules {
		handler.Handle(BaseHTTPPath+"/"+m.Name()+"/", m.HTTPHandler())
	}

	return handler
}
