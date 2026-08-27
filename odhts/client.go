// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>

// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

const RequestTimeFormat = "2006-01-02T15:04:05.000-0700"

type C struct {
	baseUrl  string
	referer  string
	tokenUrl string
	auth     *auth
	http     *http.Client
}

// StatusError reports a non-2xx response from the API.
type StatusError struct {
	StatusCode int
	URL        string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("timeseries request to %s returned status %d", e.URL, e.StatusCode)
}

/*
Initialize an Open Data Hub time series client.

	referer: identify your application to get better quota and let us know who you are
*/
func NewDefaultClient(referer string) *C {
	return NewCustomClient("https://mobility.api.opendatahub.com/v2",
		"https://auth.opendatahub.com/auth/realms/noi/protocol/openid-connect/token",
		referer)
}

/*
Initialize an Open Data Hub time series client with a custom endpoint

	baseUrl: Open Data Hub endpoint, e.b.https://mobility.api.opendatahub.com
	tokenUrl: Oauth endpoint, leave empty if not using credentials
	referer: identify your application to get better quota and let us know who you are
*/
func NewCustomClient(baseUrl string, tokenUrl string, referer string) *C {
	return &C{
		baseUrl:  baseUrl,
		tokenUrl: tokenUrl,
		referer:  referer,
		http:     &http.Client{},
	}
}

// UseHTTPClient replaces the client used for both API and token requests.
//
// There is deliberately no default timeout: a request's deadline belongs to its
// context, which every call takes. Set one here only for transport-level
// concerns a context cannot express, such as a proxy or a custom TLS config.
func (c *C) UseHTTPClient(h *http.Client) {
	if h != nil {
		c.http = h
	}
}

func (c *C) UseAuth(clientId string, clientSecret string) {
	c.auth = &auth{
		TokenUrl:     c.tokenUrl,
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
}

func ar2Path(ar []string) string {
	if len(ar) == 0 {
		return "*"
	}
	var ret string
	for i, s := range ar {
		ret += s
		if i < len(ar)-1 {
			ret += ","
		}
	}
	return ret
}

func makeHistoryPath(req *Request) string {
	return fmt.Sprintf("/%s/%s/%s/%s/%s",
		req.Repr,
		ar2Path(req.StationTypes),
		ar2Path(req.DataTypes),
		req.From.Format(RequestTimeFormat),
		req.To.Format(RequestTimeFormat))
}

func makeLatestPath(req *Request) string {
	return fmt.Sprintf("/%s/%s/%s/latest",
		req.Repr,
		ar2Path(req.StationTypes),
		ar2Path(req.DataTypes))
}

func makeStationTypePath(req *Request) string {
	return fmt.Sprintf("/%s/%s",
		req.Repr,
		ar2Path(req.StationTypes))
}

func makeQueryParam(query *url.Values, name string, value any, defaultValue any) {
	if value != defaultValue {
		query.Add(name, fmt.Sprint(value))
	}
}

func makeQuery(req *Request) *url.Values {
	query := &url.Values{}
	makeQueryParam(query, "origin", req.Origin, "")
	makeQueryParam(query, "limit", req.Limit, 200)
	makeQueryParam(query, "offset", req.Offset, 0)
	makeQueryParam(query, "select", req.Select, "")
	makeQueryParam(query, "where", req.Where, "")
	makeQueryParam(query, "shownull", req.Shownull, false)
	makeQueryParam(query, "distinct", req.Distinct, true)
	makeQueryParam(query, "timezone", req.Timezone, "")
	return query
}

func getPath[T any](ctx context.Context, c *C, path string, req *Request, result *Response[T]) error {
	if TestReqHook != nil {
		return runReqHook(req, result)
	}
	u, err := url.Parse(c.baseUrl)
	if err != nil {
		return fmt.Errorf("unable to parse Base URL from config: %w", err)
	}
	u.Path += path
	u.RawQuery = makeQuery(req).Encode()
	return c.requestUrl(ctx, u, result)
}

func StationType[T any](ctx context.Context, c *C, req *Request, res *Response[T]) error {
	return getPath(ctx, c, makeStationTypePath(req), req, res)
}

func History[T any](ctx context.Context, c *C, req *Request, res *Response[T]) error {
	return getPath(ctx, c, makeHistoryPath(req), req, res)
}

func Latest[T any](ctx context.Context, c *C, req *Request, res *Response[T]) error {
	return getPath(ctx, c, makeLatestPath(req), req, res)
}

func Get[T any](ctx context.Context, c *C, query string, result *Response[T]) error {
	u, err := url.Parse(c.baseUrl + query)
	if err != nil {
		return fmt.Errorf("unable to parse query URL: %w", err)
	}
	return c.requestUrl(ctx, u, result)
}

func (c *C) requestUrl(ctx context.Context, reqUrl *url.URL, result any) error {
	slog.Debug("Ninja request with URL: " + reqUrl.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return fmt.Errorf("unable to create Ninja HTTP Request: %w", err)
	}

	req.Header = http.Header{
		"Accept": {"application/json"},
	}

	if c.auth != nil {
		token, err := c.auth.getToken(ctx, c.httpClient())
		if err != nil {
			return fmt.Errorf("error authorizing request: %w", err)
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	if c.referer != "" {
		req.Header.Set("Referer", c.referer)
	}

	res, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("error performing ninja request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return &StatusError{StatusCode: res.StatusCode, URL: reqUrl.String()}
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	err = json.Unmarshal(bodyBytes, result)
	if err != nil {
		return fmt.Errorf("error unmarshalling response JSON to provided interface: %w", err)
	}

	return nil
}

// httpClient tolerates a zero-value C, which a caller can still construct.
func (c *C) httpClient() *http.Client {
	if c.http == nil {
		return http.DefaultClient
	}
	return c.http
}
