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

package utils

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_models "github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	pathSep     = "/"
	paramPrefix = "{"
	paramSuffix = "}"
)

func isPathParam(segment string) bool {
	return strings.HasPrefix(segment, paramPrefix) && strings.HasSuffix(segment, paramSuffix)
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func GetPathParams(specPath string, opPath string) map[string]string {
	result := make(map[string]string)

	specSegs := strings.Split(specPath, pathSep)
	opSegs := strings.Split(opPath, pathSep)

	for i, specSeg := range specSegs {
		if specSeg == opSegs[i] {
			continue
		} else if isPathParam(specSeg) {
			result[specSeg[1:len(specSeg)-1]] = opSegs[i]
		}
	}

	return result
}

func WalkFiles(root string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err == nil && d.Type().IsRegular() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return files, fmt.Errorf("error while iterating over files in '%s' directory: %w", root, err)
	}

	return files, nil
}

func ReadDictionaryFiles(filenames []string) ([]string, error) {
	lines := []string{}

	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return []string{}, fmt.Errorf("unable to read dictionary file: %s", filename)
		}
		lines = append(lines, strings.Split(string(content), "\n")...)
	}

	return lines, nil
}

func FindHeader(headers []*_models.Header, headerName string) (index int, found bool) {
	for i, header := range headers {
		if strings.EqualFold(header.Key, headerName) {
			return i, true
		}
	}
	return -1, false
}

type API = string
