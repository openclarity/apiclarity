module github.com/apiclarity/apiclarity/plugins/gateway/tyk

go 1.16

// From here: https://tyk.io/docs/plugins/supported-languages/golang/#plugin-development-flow
//nolint:gomoddirectives
replace github.com/jensneuse/graphql-go-tools => github.com/TykTechnologies/graphql-go-tools v1.6.2-0.20210324124350-140640759f4b

require (
	github.com/TykTechnologies/tyk v1.9.2-0.20210930081546-bda54b0f790c
	github.com/apiclarity/apiclarity/plugins/api v0.0.0-20211111134854-2204e564d01c
	github.com/go-openapi/runtime v0.21.0
	github.com/go-openapi/strfmt v0.21.0
)
