module github.com/apiclarity/apiclarity/backend

go 1.16

require (
	github.com/apiclarity/apiclarity/api v0.0.0
	github.com/apiclarity/apiclarity/plugins/api v0.0.0
	github.com/apiclarity/speculator v0.0.5-0.20211007151558-df7e428069c8
	github.com/bxcodec/faker/v3 v3.6.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/loads v0.21.0
	github.com/go-openapi/runtime v0.21.0
	github.com/go-openapi/spec v0.20.4
	github.com/go-openapi/strfmt v0.21.0
	github.com/go-openapi/validate v0.20.3
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/rs/cors v1.8.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	github.com/urfave/cli v1.22.5
	gorm.io/driver/postgres v1.1.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.15
	gotest.tools/v3 v3.0.3
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/apiclarity/apiclarity/api v0.0.0 => ./../api

replace github.com/apiclarity/apiclarity/plugins/api v0.0.0 => ./../plugins/api
