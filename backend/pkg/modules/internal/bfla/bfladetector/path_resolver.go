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

package bfladetector

import (
	"context"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func NewPathResolver(provider OpenAPIProvider) PathResolver {
	return &pathResolver{openapiProvider: provider}
}

type PathResolver interface {
	RezolvePath(ctx context.Context, apiID uint, uri string) (urlpath string, specType SpecType)
}

type pathResolver struct {
	openapiProvider OpenAPIProvider
}

func (r *pathResolver) RezolvePath(ctx context.Context, apiID uint, uri string) (urlpath string, specType SpecType) {
	defer func() {
		log.Infof("path: %s; resolved path: %s", uri, urlpath)
	}()
	u, _ := url.Parse(uri)
	urlpath = u.Path
	spec, specType := r.getServiceOpenapiSpec(ctx, apiID)
	if spec != nil {
		pathDef, _ := r.matchSpecAndPath(u.Path, spec)
		if pathDef != "" {
			return pathDef, specType
		}
	}
	return urlpath, specType
}

func (r *pathResolver) getServiceOpenapiSpec(ctx context.Context, apiID uint) (*GenericOpenapiSpec, SpecType) {
	reader, specType, err := r.openapiProvider.GetOpenAPI(ctx, apiID)
	if err != nil {
		log.Error("unable to get openapi spec: ", err)
		return nil, specType
	}
	s := &GenericOpenapiSpec{}
	if err := yaml.NewDecoder(reader).Decode(s); err != nil {
		log.Error(err)
		return nil, specType
	}
	return s, specType
}

func (r *pathResolver) matchSpecAndPath(path string, spec *GenericOpenapiSpec) (pathDef string, paramValues map[string]string) {
	pathSplit := strings.Split(path, "/")
pathsLoop:
	for pathDefKey, pathItem := range spec.Paths {
		params := map[string]string{}
		for _, param := range pathItem.Parameters {
			if param.In == "path" {
				params[param.Name] = ""
			}
		}
		for _, op := range pathItem.Operations {
			for _, param := range op.Parameters {
				if param.In == "path" {
					params[param.Name] = ""
				}
			}
		}

		pathDefSplit := strings.Split(pathDefKey, "/")
		if len(pathDefSplit) != len(pathSplit) {
			continue
		}
		for i := range pathDefSplit {
			if pathDefSplit[i] == pathSplit[i] {
				if i == len(pathSplit)-1 {
					return pathDefKey, params
				}
				continue
			}
			pathPart := strings.TrimLeft(strings.TrimRight(pathDefSplit[i], "}"), "{")
			if _, ok := params[pathPart]; ok && strings.HasSuffix(pathDefSplit[i], "}") && strings.HasPrefix(pathDefSplit[i], "{") {
				params[pathPart] = pathSplit[i]
				if i == len(pathSplit)-1 {
					return pathDefKey, params
				}
				continue
			}
			continue pathsLoop
		}
	}
	return "", nil
}
