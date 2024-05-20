module github.com/openclarity/apiclarity/plugins/gateway/kong

go 1.19

require (
	github.com/Kong/go-pdk v0.10.0
	github.com/go-openapi/strfmt v0.21.7
	github.com/openclarity/apiclarity/plugins/api v0.0.0
	github.com/openclarity/apiclarity/plugins/common v0.0.0
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/go-openapi/analysis v0.20.1 // indirect
	github.com/go-openapi/errors v0.20.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/loads v0.21.0 // indirect
	github.com/go-openapi/runtime v0.21.0 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.20.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/openclarity/trace-sampling-manager/api v0.0.0-20220503091058-2730f5ebad2c // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	go.mongodb.org/mongo-driver v1.11.3 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/openclarity/apiclarity/plugins/api v0.0.0 => ./../../api

replace github.com/openclarity/apiclarity/plugins/common v0.0.0 => ./../../common
