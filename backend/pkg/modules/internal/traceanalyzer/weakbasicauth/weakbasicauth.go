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

package weakbasicauth

import (
	"bytes"
	"encoding/base64"
	"sort"
	"strings"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	AuthorizationHeader = "authorization"
	BasicAuth           = "Basic"
	ShortPasswordLen    = 8
)

// Extracts the Basic Authentication token from the Query.
func findBasicAuthToken(trace *models.Telemetry) (user, password string, found bool) {
	index, found := utils.FindHeader(trace.Request.Common.Headers, AuthorizationHeader)
	if !found {
		return "", "", false
	}

	hValue := trace.Request.Common.Headers[index].Value
	splitValue := strings.Split(hValue, " ")
	if len(splitValue) != 2 || splitValue[0] != BasicAuth {
		return "", "", false
	}

	decodedAuth, err := base64.StdEncoding.DecodeString(splitValue[1])
	if err != nil {
		return "", "", false
	}
	splitAuth := bytes.Split(decodedAuth, []byte{':'})
	//nolint:gomnd
	if len(splitAuth) != 2 {
		return "", "", false
	}

	return string(splitAuth[0]), string(splitAuth[1]), true
}

type userPassword struct {
	User     string
	Password string
}

type WeakBasicAuth struct {
	shortPasswordLen int
	knownPasswordsAC ahocorasick.AhoCorasick
	usedCredentials  map[userPassword]map[utils.API]bool
}

func NewWeakBasicAuth(knownPasswords []string) *WeakBasicAuth {
	acBuilder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: false,
		MatchOnlyWholeWords:  true,
		MatchKind:            ahocorasick.LeftMostLongestMatch,
		DFA:                  true,
	})

	return &WeakBasicAuth{
		shortPasswordLen: ShortPasswordLen,
		knownPasswordsAC: acBuilder.Build(knownPasswords),
		usedCredentials:  make(map[userPassword]map[utils.API]bool),
	}
}

func (w *WeakBasicAuth) analyzeShortPassword(password string) (anns []utils.TraceAnalyzerAnnotation) {
	if len(password) <= w.shortPasswordLen {
		a := NewAnnotationShortPassword(password, w.shortPasswordLen)
		anns = append(anns, a)
	}

	return anns
}

func (w *WeakBasicAuth) analyzeKnownPassword(password string) (anns []utils.TraceAnalyzerAnnotation) {
	matches := w.knownPasswordsAC.FindAll(password)
	if len(matches) > 0 {
		anns = append(anns, NewAnnotationKnownPassword(password))
	}

	return anns
}

func (w *WeakBasicAuth) analyzeSameCreds(api utils.API, user string, password string) (anns []utils.TraceAnalyzerAnnotation) {
	up := userPassword{user, password}

	apis, ok := w.usedCredentials[up]
	if !ok {
		// Nobody else is using this user/password, create a new entry
		w.usedCredentials[up] = make(map[utils.API]bool)
		w.usedCredentials[up][api] = true
	} else if !apis[api] {
		// There is already at least one other Api using this user/password
		w.usedCredentials[up][api] = true
		listOfAPIs := []string{}
		for sameAPI := range w.usedCredentials[up] {
			listOfAPIs = append(listOfAPIs, sameAPI)
		}
		sort.Strings(listOfAPIs)
		anns = append(anns, NewAnnotationSamePassword(user, password, listOfAPIs))
	}
	// Else this api was already added here, no need to report an observation.

	return anns
}

func (w *WeakBasicAuth) Analyze(trace *models.Telemetry) (eventAnns []utils.TraceAnalyzerAnnotation) {
	api := trace.Request.Host

	user, password, found := findBasicAuthToken(trace)
	if found {
		eventAnns = append(eventAnns, w.analyzeShortPassword(password)...)
		eventAnns = append(eventAnns, w.analyzeKnownPassword(password)...)
		eventAnns = append(eventAnns, w.analyzeSameCreds(api, user, password)...)
	}

	return eventAnns
}
