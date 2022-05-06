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
	"fmt"
	"sort"
	"strconv"
	"strings"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	"github.com/openclarity/apiclarity/plugins/api/server/models"
)

const (
	AuthorizationHeader = "authorization"
	BasicAuth           = "Basic"
	ShortPasswordLen    = 8
)

const (
	KindShortPassword = "BASIC_AUTH_SHORT_PASSWORD"
	KindKnownPassword = "BASIC_AUTH_KNOWN_PASSWORD"
	KindSamePassword  = "BASIC_AUTH_SAME_PASSWORD"
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
	usedCredentials  map[userPassword]map[utils.Api]bool
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
		usedCredentials:  make(map[userPassword]map[utils.Api]bool),
	}
}

func (w *WeakBasicAuth) analyzeShortPassword(password string) (anns []core.Annotation) {
	if len(password) <= w.shortPasswordLen {
		anns = append(anns, core.Annotation{
			Name:       KindShortPassword,
			Annotation: []byte(strconv.Itoa(len(password))),
		})
	}

	return anns
}

func (w *WeakBasicAuth) analyzeKnownPassword(password string) (anns []core.Annotation) {
	matches := w.knownPasswordsAC.FindAll(password)
	if len(matches) > 0 {
		anns = append(anns, core.Annotation{
			Name:       KindKnownPassword,
			Annotation: []byte(password),
		})
	}

	return anns
}

func (w *WeakBasicAuth) analyzeSameCreds(api utils.Api, user string, password string) (anns []core.Annotation) {
	up := userPassword{user, password}

	apis, ok := w.usedCredentials[up]
	if !ok {
		// Nobody else is using this user/password, create a new entry
		w.usedCredentials[up] = make(map[utils.Api]bool)
		w.usedCredentials[up][api] = true
	} else if !apis[api] {
		// There is already at least one other Api using this user/password
		w.usedCredentials[up][api] = true
		listOfAPIs := []string{}
		for sameAPI := range w.usedCredentials[up] {
			listOfAPIs = append(listOfAPIs, sameAPI)
		}
		sort.Strings(listOfAPIs)
		anns = append(anns, core.Annotation{
			Name:       KindSamePassword,
			Annotation: []byte(fmt.Sprintf("%s:%s,%s", user, password, strings.Join(listOfAPIs, ","))),
		})
	} else {
		// This api was already added here, no need to report an observation
	}

	return anns
}

func (w *WeakBasicAuth) Analyze(trace *models.Telemetry) (eventAnns []core.Annotation, apiAnns []core.Annotation) {
	api := trace.Request.Host

	user, password, found := findBasicAuthToken(trace)
	if found {
		eventAnns = append(eventAnns, w.analyzeShortPassword(password)...)
		eventAnns = append(eventAnns, w.analyzeKnownPassword(password)...)
		eventAnns = append(eventAnns, w.analyzeSameCreds(api, user, password)...)
	}

	return eventAnns, apiAnns
}
