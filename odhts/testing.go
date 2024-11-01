// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"encoding/json"
	"os"
	"reflect"
)

// For testing purposes, set this function and it will be used to retrieve a request instead of http
var TestReqHook func(*Request) (any, error)

func runReqHook(req *Request, result any) error {
	r, err := TestReqHook(req)
	if err != nil {
		return err
	}
	// unholy memcpy: sets the memory at result to the value of r.

	// They have to be the same type
	// Has to be this way, because golang does not allow for variables to hold parametrized functions.
	// So we have to use 'any' and hack around it with reflection. The golang.json lib does the same
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(r).Elem())
	return nil
}

func LoadJsonFile[T any](f string) (*Response[T], error) {
	r := &Response[T]{}
	b, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(b, r)
	return r, nil
}
