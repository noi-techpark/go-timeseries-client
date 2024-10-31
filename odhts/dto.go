// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"strings"
	"time"
)

// Top level response object
// The user can supply a custom data structure that implements json.Unmarshaler
type Response[Dtype any] struct {
	Data   Dtype  `json:"data"`
	Offset uint64 `json:"offset"`
	Limit  int64  `json:"limit"`
}

// Time format used by timeseries API, implements json.Unmarshaler
type TsTime struct {
	time.Time
}

// Convenience DTO for typical station requests
type StationDto[T any] struct {
	Scode   string
	Sname   string
	Sorigin string
	Scoord  CoordDto `json:"scoordinate"`
	Smeta   T        `json:"smetadata"`
}

type CoordDto struct {
	X    float32
	Y    float32
	Srid uint32
}

// Convenience DTO for typical /latest requests
type LatestDto struct {
	MPeriod    int    `json:"mperiod"`
	MValidTime TsTime `json:"mvalidtime"`
	MValue     int    `json:"mvalue"`
	Scode      string `json:"scode"`
	Stype      string `json:"stype"`
	Tname      string `json:"tname"`
}

func (nt *TsTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		nt.Time = time.Time{}
		return
	}
	nt.Time, _ = time.Parse("2006-01-02 15:04:05.000-0700", s)
	return
}
