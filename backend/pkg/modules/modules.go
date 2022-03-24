package modules

import (
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"

	_ "github.com/apiclarity/apiclarity/backend/pkg/modules/internal/demo"
)

type (
	Module          = core.Module
	Annotation      = core.Annotation
	BackendAccessor = core.BackendAccessor
	Event           = core.Event
)

var New = core.New
