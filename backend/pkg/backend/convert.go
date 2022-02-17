package backend

import (
	"github.com/apiclarity/apiclarity/plugins/api/server/models"
	"github.com/apiclarity/speculator/pkg/spec"
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
