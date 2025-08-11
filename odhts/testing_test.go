// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"testing"

	"gotest.tools/v3/assert"
)

type BikeShareMeta struct {
	Address   string
	TotalBays int `json:"total-bays"`
}

func TestLoadJson(t *testing.T) {
	j, err := LoadJsonFile[[]StationDto[BikeShareMeta]]("test/ninja_loadjson.json")
	assert.NilError(t, err, "Failed to load JSON")
	assert.Equal(t, j.Data[0].Sname, "Viale della Stazione - Bahnhofsallee", "Unexpected mapping from JSON")
	assert.Equal(t, j.Data[0].Smeta.TotalBays, 12, "Unexpected mapping from JSON")
}

func TestReqHook1(t *testing.T) {
	j, err := LoadJsonFile[[]StationDto[BikeShareMeta]]("test/ninja_loadjson.json")
	assert.NilError(t, err, "Failed to load JSON")

	c := NewDefaultClient("")
	req := Request{}
	req.Origin = "test"
	TestReqHook = func(nr *Request) (any, error) {
		assert.Equal(t, nr.Origin, req.Origin, "Passed request not matching the one in hook")
		return j, nil
	}

	res := Response[[]StationDto[BikeShareMeta]]{}
	err = StationType(c, &req, &res)
	assert.NilError(t, err, "Error calling ninja with req hook")

	assert.Assert(t, res.Data[0].Scode != "", "zero value in returned data")
	assert.Equal(t, j.Data[0].Scode, res.Data[0].Scode, "Mismatch between returned value and hook")
}
