// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package global

//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -old-config-style -generate chi-server,client,types,spec,skip-prune -package global  -o restapi.gen.go --import-mapping=../common/openapi.yaml:github.com/openclarity/apiclarity/api3/common openapi.gen.yaml
