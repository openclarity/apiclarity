package restapi

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -old-config-style -generate chi-server,types,spec,skip-prune -package restapi -o restapi.gen.go --import-mapping=../../../../../../api3/common/openapi.yaml:github.com/openclarity/apiclarity/api3/common openapi.yaml
