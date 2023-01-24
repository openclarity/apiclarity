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

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	logging "github.com/sirupsen/logrus"

	globalapi "github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
)

var staticFakeTest = `
{
	"steps": [
		{"report":{},"progress":0,"status":"IN_PROGRESS"},
		{"report":{},"progress": 100,"status":"DONE"}
	]
}
`

type FakeTest struct {
	Steps []globalapi.FuzzingStatusAndReport `json:"steps"`
}

type FakeClient struct {
	testFileName string
	remoteHost   string
	quitPipe     chan bool
}

func (c *FakeClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Infof("[Fuzzer][FakeClient] TriggerFuzzingJob(%v, %v, %v, %v):: --> <--", apiID, endpoint, securityItem, timeBudget)
	go FakeTriggerFuzzingJob(context.TODO(), c.quitPipe, c.testFileName, uint(apiID), c.remoteHost)
	return nil
}

func (c *FakeClient) StopFuzzingJob(apiID int64, complete bool) error {
	// Just quit goroutine
	if !complete {
		c.quitPipe <- true
	}
	return nil
}

func SendReport(ctx context.Context, client *globalapi.Client, apiID uint, report globalapi.FuzzingStatusAndReport) error {
	resp, err := client.FuzzerPostUpdateStatus(ctx, int64(apiID), report)
	if err != nil {
		return fmt.Errorf("send report error: %v", err)
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("send report error: got unexpected response code (%v)", resp.StatusCode)
	}
	return nil
}

func flushChannel(pipe chan bool) {
	for len(pipe) > 0 {
		<-pipe
	}
}

func FakeTriggerFuzzingJob(ctx context.Context, pipe chan bool, testFilename string, apiID uint, remoteHost string) {
	flushChannel(pipe)

	var testBytes []byte
	testFile, err := os.Open(testFilename)
	if err == nil {
		logging.Infof("[Fuzzer][FakeClient] Use data from (%v)", testFilename)
		defer testFile.Close()

		testBytes, _ = ioutil.ReadAll(testFile)
	} else {
		testBytes = []byte(staticFakeTest)
	}

	fakeTest := FakeTest{}
	err = json.Unmarshal(testBytes, &fakeTest)
	if err != nil {
		logging.Errorf("can't load test file (%v): %v", testFilename, err)
	}
	logging.Errorf("can't read file (%v): %v", testFilename, err)
	apicClient, err := globalapi.NewClient(remoteHost)
	if err != nil {
		logging.Errorf("[Fuzzer][FakeClient] unable to connect to APIClarity: %v", err)
	}
	for _, step := range fakeTest.Steps {
		select {
		case <-pipe:
			logging.Infof("[Fuzzer][FakeClient] Interrupt the test")
			return
		default:
			logging.Infof("[Fuzzer][FakeClient] inject data %v", step)
			err = SendReport(ctx, apicClient, apiID, step)
			if err != nil {
				logging.Errorf("Failed to send report to (%v): %v", remoteHost, err)
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// nolint: ireturn,nolintlint
func NewFakeClient(config *config.Config) (Client, error) {
	p := &FakeClient{
		testFileName: config.GetFakeFileName(),
		remoteHost:   config.GetPlatformHost(),
		quitPipe:     make(chan bool),
	}
	return p, nil
}
