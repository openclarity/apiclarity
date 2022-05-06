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
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestCompression(t *testing.T) {
}

type TestCase struct {
	v          []string
	maxHistory int
	guessable  bool
}

func TestHistory(t *testing.T) {
	testcases := []TestCase{
		{
			v:         []string{"0000001", "0000002", "0000003", "0000004", "0000005", "0000006", "0000007", "0000008", "0000009", "0000010"},
			guessable: true,
		},
		{
			v:          []string{"124", "125", "126", "123", "12aze3", "1sqdf23", "1yy23", "1ztr23", "123", "123", "1zertezrt23", "123", "123", "123", "123", "123", "123", "123", "123", "123"},
			maxHistory: 200,
			guessable:  false, // Not enough data
		},
		{
			v:         []string{"124", "125", "126", "123", "12aze3", "1sqdf23", "1yy23", "1ztr23", "123", "123", "1zertezrt23", "123", "123", "123", "123", "123", "123", "123", "123", "123"},
			guessable: false, // Not enough data
		},
		{
			v:         []string{"123", "123", "123", "123", "123", "123", "123", "123", "123", "123"},
			guessable: false, // Not enough different data
		},
		{
			v: []string{
				"user0001", "user0002", "user0003", "user0004", "user0005",
				"user9991", "user9992", "user9993", "user9994", "user9995", "user9996",
				"user8002", "user8003", "user8004", "user8005", "user8006", "user8007",
				"userA00D", "userC00F", "userA00X", "userI00L",
			},
			guessable: false,
		},
		{
			v: []string{
				"user0001", "user0002", "user0003", "user0004", "user0005",
				"user8002", "user8003", "user8004", "user8005", "user8006", "user8007",
			},
			guessable: true,
		},
		{
			v:         []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"},
			guessable: false,
		},
		{
			v:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "ja"},
			guessable: false,
		},
		{
			v:         []string{"Paris", "Londres", "New-York", "Madrid", "Roma", "Milan", "Marseille", "San-Jose", "Prapoutel les 7 laux", "Tokyo"},
			guessable: false,
		},
	}

	nbUUID := 100
	uuids := make([]string, nbUUID)
	for i := 0; i < nbUUID; i++ {
		u := uuid.NewV4()
		uuids[i] = u.String()
	}

	testcases = append(testcases, TestCase{
		v:          uuids,
		maxHistory: nbUUID,
		guessable:  false,
	})

	for _, tc := range testcases {
		var maxHistory int
		if tc.maxHistory != 0 {
			maxHistory = tc.maxHistory
		} else {
			maxHistory = len(tc.v)
		}
		ga := NewGuessableAnalyzer(uint(maxHistory))
		var r bool
		for _, s := range tc.v {
			r, _ = ga.IsGuessableParam("/pet/{petId}", "petId", s)
		}

		if r != tc.guessable {
			t.Errorf("Wanted: %v, got: %v (%+v)", tc.guessable, r, tc)
		}
	}
}
