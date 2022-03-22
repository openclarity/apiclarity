package modules

import (
	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/core"
)

type (
	Module          = core.Module
	Annotation      = core.Annotation
	BackendAccessor = core.BackendAccessor
	Event           = core.Event
)

var New = core.New
