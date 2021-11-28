module github.com/apiclarity/apiclarity/plugins/gateway/tyk

go 1.16

// From here: https://tyk.io/docs/plugins/supported-languages/golang/#plugin-development-flow
//nolint:gomoddirectives
replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20211112130051-ad1e36a78a9a

require (
	github.com/TykTechnologies/tyk v1.9.2-0.20211119141645-a642669fba58
	github.com/apiclarity/apiclarity/plugins/api v0.0.0
	github.com/gin-gonic/gin v1.7.0 // indirect
	github.com/go-openapi/runtime v0.21.0
	github.com/go-openapi/strfmt v0.21.0
)

replace github.com/apiclarity/apiclarity/plugins/api v0.0.0 => ./../../../api
