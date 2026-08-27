// SPDX-FileCopyrightText: 2024 NOI Techpark <digital@noi.bz.it>
//
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
	"strconv"
	"strings"
	"sync"
	"time"
)

type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	NotBeforePolicy  int64  `json:"not-before-policy"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string
}

type auth struct {
	TokenUrl     string
	ClientId     string
	ClientSecret string

	// A client is shared by value but its auth is not, so the cached token is
	// reachable from every copy and from every goroutine using them.
	mu          sync.Mutex
	token       Token
	tokenExpiry int64
}

func (a *auth) getToken(ctx context.Context, h *http.Client) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.token.AccessToken == "" || time.Now().Unix() > a.tokenExpiry {
		if err := a.newToken(ctx, h); err != nil {
			return "", err
		}
	}
	return a.token.AccessToken, nil
}

func (a *auth) newToken(ctx context.Context, h *http.Client) error {
	slog.Debug("requesting a new timeseries token")
	params := url.Values{}
	params.Add("client_id", a.ClientId)
	params.Add("client_secret", a.ClientSecret)
	params.Add("grant_type", "client_credentials")

	return a.authRequest(ctx, h, params)
}

func (a *auth) authRequest(ctx context.Context, h *http.Client, params url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.TokenUrl,
		strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("unable to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	// A non-OK response used to fall through silently, leaving the token empty
	// and sending "Authorization: Bearer " on every subsequent request. Wrong
	// credentials then presented as missing data rather than as an error.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token endpoint %s returned status %d", a.TokenUrl, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read token response: %w", err)
	}
	if err := json.Unmarshal(body, &a.token); err != nil {
		return fmt.Errorf("unable to decode token response: %w", err)
	}

	// Expire early so a token is never used in the last moments of its life.
	a.tokenExpiry = time.Now().Unix() + a.token.ExpiresIn - 600
	slog.Debug("timeseries token expires at " + strconv.FormatInt(a.tokenExpiry, 10))
	return nil
}
