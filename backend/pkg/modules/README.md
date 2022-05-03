### Create a new module

1. Create a new module create a go package under `backend/pkg/modules/internal`
2. Implement the `backend/pkg/modules/internal/core.Module` interface
3. Create an init function and call the `core.RegisterModule` function that should do the initialization stems and return the module
4. Create a `openapi.yaml` for the http handlers and generate the code using oapi-codegen
   1. Tip: create a `gen.go` file and add a `//go:generate` annotation for generating the server easier
5. Add the import to the new module with underscore. Eg: `_ "github.com/openclarity/apiclarity/backend/pkg/modules/internal/demo"`

### Create alerts
Alerts are a way to annotate an event to signal that it has issues.
The alerts can be of 2 types: `ALERT_INFO` and `ALERT_WARN`

The alerts are just annotations and should be treated as such 
the core has the predefined alerts `core.AlertInfoAnn` and `core.AlertWarnAnn`

Example setting Alerts:
```go
accessor.CreateAPIEventAnnotations(ctx, "module_name", 1, core.AlertInfoAnn)
```