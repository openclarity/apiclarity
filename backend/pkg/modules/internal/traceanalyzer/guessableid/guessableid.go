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

package guessableid

import (
	"bytes"
	"compress/gzip"
	"strings"

	edlib "github.com/hbollon/go-edlib"
	uuid "github.com/satori/go.uuid"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

const gzipHeaderLen = 23

const (
	MaxParamHistory = 10

	HintTypeThreshold    = 0.8 // When guessing the type, the dataset ratio need to be at least this value
	DistanceThreshold    = 0.8 // Close to 1 means very similar
	CompressionThreshold = 2.0 // High means better compression, means very similar
)

type paramLocKey struct {
	operation, name string
}

type TypeHint int

const (
	HintUnknown TypeHint = iota
	HintUUID
	HintInteger
)

type paramHistory struct {
	maxHistory uint
	history    []string
	enoughData bool
	i          int
	typeHint   TypeHint
}

func newParamHistory(maxHistory uint) *paramHistory {
	return &paramHistory{
		maxHistory: maxHistory,
		history:    make([]string, maxHistory),
	}
}

func (p *paramHistory) add(value string) bool {
	// Do not add it if already present
	for _, v := range p.history {
		if v == value {
			return false // We did not add it
		}
	}
	p.history[p.i] = value
	p.i = (p.i + 1) % int(p.maxHistory)

	// We reached the end of the array, and circled back to the start.
	// We have enough data to do something
	// Let's at least try to guess what is the type of this data
	if p.i == 0 {
		p.enoughData = true
		p.guessTypeHint()
	}

	return true
}

func (p *paramHistory) guessTypeHint() {
	if !p.enoughData { // Should not happen
		return
	}

	typeHintHist := make(map[TypeHint]uint)
	for _, param := range p.history {
		// Check if it's a UUID
		if _, err := uuid.FromString(param); err == nil {
			typeHintHist[HintUUID]++
		} else {
			typeHintHist[HintUnknown]++
		}
	}

	// Get the most represented typeHint
	max := uint(0)
	typeHint := HintUnknown
	for k, v := range typeHintHist {
		if v > max {
			max = v
			typeHint = k
		}
	}

	if float64(max)/float64(p.maxHistory) > HintTypeThreshold {
		p.typeHint = typeHint
	}
}

func (p *paramHistory) isSimilar() (bool, GuessableReason) {
	if p.typeHint == HintUUID { // It looks like this is a UUID parameter, let's consider it as non guessable
		return false, GuessableReason{}
	}
	d := p.distance()
	c := p.compressionRatio()

	// Heuristic of the death. If the distance is close to 1, AND the compression
	// ratio is "pretty" good, it means that the set is pretty similar
	if d >= DistanceThreshold && c >= CompressionThreshold {
		return true, GuessableReason{
			Distance:             d,
			DistanceThreshold:    DistanceThreshold,
			CompressionRatio:     c,
			CompressionThreshold: CompressionThreshold,
		}
	}

	return false, GuessableReason{}
}

func (p *paramHistory) compressionRatio() float32 {
	var result float32

	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return result
	}

	b := strings.Join(p.history[:], "")
	_, err = zw.Write([]byte(b))
	if err != nil {
		return result
	}
	if err := zw.Close(); err != nil {
		return result
	}

	compressedLen := buf.Len() - gzipHeaderLen

	return float32(len(b)) / float32(compressedLen)
}

func (p *paramHistory) distance() float32 {
	sMatrix := make([][]float32, p.maxHistory)
	for i := range sMatrix {
		sMatrix[i] = make([]float32, p.maxHistory) // TODO: Optimize creation the triangular matrix should be pre instantiated in the paramHistory struct
	}
	for i, pi := range p.history {
		for j, pj := range p.history {
			if i == j {
				// Only compute inferior triangular part
				break
			}
			res, err := edlib.StringsSimilarity(pi, pj, edlib.OSADamerauLevenshtein)
			if err != nil {
				sMatrix[i][j] = 0.0
			} else {
				sMatrix[i][j] = res
			}
		}
	}

	var sum float32
	for i := range sMatrix {
		for j := range sMatrix[i] {
			if i == j {
				// Only compute inferior triangular part
				break
			}
			sum += sMatrix[i][j]
		}
	}
	//nolint:gomnd
	nbLowerMatrix := float64(p.maxHistory) * (float64(p.maxHistory) - 1.0) / 2.0
	ratioToMax := float64(sum) / nbLowerMatrix

	return float32(ratioToMax)
}

type GuessableAnalyzer struct {
	maxHistory uint
	history    map[paramLocKey]*paramHistory
}

func NewGuessableAnalyzer(maxHistory uint) *GuessableAnalyzer {
	return &GuessableAnalyzer{
		maxHistory: maxHistory,
		history:    make(map[paramLocKey]*paramHistory),
	}
}

func (g *GuessableAnalyzer) IsGuessableParam(location string, name string, value string) (bool, GuessableReason) {
	key := paramLocKey{location, name}
	timeToCheck := g.learnParam(key, value)

	if timeToCheck && g.history[key].enoughData {
		if similar, reason := g.history[key].isSimilar(); similar {
			return true, reason
		}
	}

	return false, GuessableReason{}
}

func (g *GuessableAnalyzer) learnParam(key paramLocKey, value string) bool {
	p := g.history[key]
	if p == nil {
		p = newParamHistory(g.maxHistory)
		g.history[key] = p
	}

	p.add(value)

	return p.i == 0
}

func (g *GuessableAnalyzer) Analyze(path, method string, pathParams map[string]string, trace *pluginsmodels.Telemetry) (eventAnns []utils.TraceAnalyzerAnnotation, apiAnns []utils.TraceAnalyzerAPIAnnotation) {
	guessableParams := []GuessableParameter{}

	for pName, pValue := range pathParams {
		if isGuessable, reason := g.IsGuessableParam(path, pName, pValue); isGuessable {
			guessableParams = append(guessableParams, GuessableParameter{Name: pName, Value: pValue, Reason: reason})
		}
	}

	if len(guessableParams) > 0 {
		eventAnns = append(eventAnns, NewAnnotationGuessableID(path, method, guessableParams))
	}

	return
}
