package utils

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"gotest.tools/assert"

	"github.com/apiclarity/apiclarity/api/client/client"
	"github.com/apiclarity/apiclarity/api/client/client/operations"
	"github.com/apiclarity/apiclarity/api/client/models"
)

func AssertGetAPIInventory(t *testing.T, apiclarityAPI *client.APIClarityAPIs, want *operations.GetAPIInventoryOKBody) {
	params := operations.NewGetAPIInventoryParams().WithPage(0).WithPageSize(50).WithType(string(models.APITypeINTERNAL)).WithSortKey("name")
	res, err := apiclarityAPI.Operations.GetAPIInventory(params)
	assert.NilError(t, err)
	assert.DeepEqual(t, res.Payload, want)
}

func AssertGetAPIEvents(t *testing.T, apiclarityAPI *client.APIClarityAPIs, want *operations.GetAPIEventsOKBody) {
	startTime, _ := time.Parse("2006-01-02T15:04:05.000Z", "2021-04-26T11:35:49.775Z")
	endTime, _ := time.Parse("2006-01-02T15:04:05.000Z", "2030-04-26T11:35:49.775Z")

	params := operations.NewGetAPIEventsParams().WithPage(0).WithPageSize(50).WithStartTime(strfmt.DateTime(startTime)).WithEndTime(strfmt.DateTime(endTime)).WithSortKey("time").WithShowNonAPI(false)
	res, err := apiclarityAPI.Operations.GetAPIEvents(params)
	assert.NilError(t, err)
	assert.DeepEqual(t, res.Payload, want)
}
