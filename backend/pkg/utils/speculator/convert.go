// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package speculator

import (
	"github.com/openclarity/apiclarity/plugins/api/server/models"
	"github.com/openclarity/speculator/pkg/spec"
)

func ConvertModelsToSpeculatorTelemetry(telemetry *models.Telemetry) *spec.Telemetry {
	return &spec.Telemetry{
		DestinationAddress:   telemetry.DestinationAddress,
		DestinationNamespace: telemetry.DestinationNamespace,
		Request:              convertRequest(telemetry.Request),
		RequestID:            telemetry.RequestID,
		Response:             convertResponse(telemetry.Response),
		Scheme:               telemetry.Scheme,
		SourceAddress:        telemetry.SourceAddress,
	}
}

func convertRequest(request *models.Request) *spec.Request {
	return &spec.Request{
		Common: convertCommon(request.Common),
		Host:   request.Host,
		Method: request.Method,
		Path:   request.Path,
	}
}

func convertResponse(response *models.Response) *spec.Response {
	return &spec.Response{
		Common:     convertCommon(response.Common),
		StatusCode: response.StatusCode,
	}
}

func convertCommon(common *models.Common) *spec.Common {
	return &spec.Common{
		TruncatedBody: common.TruncatedBody,
		Body:          common.Body,
		Headers:       convertHeaders(common.Headers),
		Version:       common.Version,
	}
}

func convertHeaders(headers []*models.Header) []*spec.Header {
	var ret []*spec.Header

	for _, header := range headers {
		ret = append(ret, &spec.Header{
			Key:   header.Key,
			Value: header.Value,
		})
	}
	return ret
}
