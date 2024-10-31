// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package ninja

import (
	"strings"
	"time"
)

type NinjaResponse[Dtype any] struct {
	Data   Dtype  `json:"data"`
	Offset uint64 `json:"offset"`
	Limit  int64  `json:"limit"`
}

type NinjaTime struct {
	time.Time
}

func (nt *NinjaTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		nt.Time = time.Time{}
		return
	}
	nt.Time, _ = time.Parse("2006-01-02 15:04:05.000-0700", s)
	return
}
