package demo

import (
	"context"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/apiclarity/apiclarity/plugins/api/server/models"
	"net/http"
)

func init() {
	core.RegisterModule(newModule)
}

func newModule(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	accessor.CreateAPIEventAnnotations(ctx, "mod", 1, core.AlertInfoAnn)
	return &demo{}, nil
}

type demo struct {
}

func (d *demo) Name() string {
	//TODO implement me
	panic("implement me")
}

func (d *demo) EventNotify(event *database.APIEvent, trace *models.Telemetry) {
	//TODO implement me
	panic("implement me")
}

func (d *demo) HTTPHandler() http.Handler {
	//TODO implement me
	panic("implement me")
}
