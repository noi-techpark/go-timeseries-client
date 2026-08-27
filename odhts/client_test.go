// SPDX-FileCopyrightText: NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: MPL-2.0

package odhts

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// stub serves both the token endpoint and the API, counting each.
func stub(t *testing.T) (*httptest.Server, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	var tokens, apis atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			tokens.Add(1)
			fmt.Fprint(w, `{"access_token":"tok","expires_in":3600,"token_type":"Bearer"}`)
			return
		}
		apis.Add(1)
		fmt.Fprint(w, `{"data":[],"offset":0,"limit":-1}`)
	}))
	t.Cleanup(s.Close)
	return s, &tokens, &apis
}

// The client is passed by value to every request function, so a token cached
// on a copy is discarded. Every call then re-authenticated.
func TestTokenIsCachedAcrossCalls(t *testing.T) {
	srv, tokens, _ := stub(t)
	c := NewCustomClient(srv.URL, srv.URL+"/token", "test")
	c.UseAuth("id", "secret")

	for i := 0; i < 5; i++ {
		var res Response[[]map[string]any]
		if err := StationType(context.Background(), c, DefaultRequest(), &res); err != nil {
			t.Fatal(err)
		}
	}
	if got := tokens.Load(); got != 1 {
		t.Errorf("token fetched %d times for 5 requests, want 1", got)
	}
}

// Sharing the token makes it reachable from several goroutines at once, which
// the per-copy version never was. Run with -race.
func TestConcurrentRequestsShareOneToken(t *testing.T) {
	srv, tokens, apis := stub(t)
	c := NewCustomClient(srv.URL, srv.URL+"/token", "test")
	c.UseAuth("id", "secret")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var res Response[[]map[string]any]
			if err := Latest(context.Background(), c, DefaultRequest(), &res); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if apis.Load() != 20 {
		t.Errorf("api calls = %d, want 20", apis.Load())
	}
	if got := tokens.Load(); got != 1 {
		t.Errorf("token fetched %d times concurrently, want 1", got)
	}
}

// Auth is optional and most callers never set it.
func TestAnonymousClient(t *testing.T) {
	srv, tokens, apis := stub(t)
	c := NewCustomClient(srv.URL, srv.URL+"/token", "test")

	var res Response[[]map[string]any]
	if err := StationType(context.Background(), c, DefaultRequest(), &res); err != nil {
		t.Fatalf("anonymous request failed: %v", err)
	}
	if apis.Load() != 1 || tokens.Load() != 0 {
		t.Errorf("api=%d token=%d, want 1 and 0", apis.Load(), tokens.Load())
	}
}

// A cancelled context must abandon the request rather than run to completion.
func TestContextCancels(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer slow.Close()

	c := NewCustomClient(slow.URL, "", "test")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	var res Response[[]map[string]any]
	err := StationType(ctx, c, DefaultRequest(), &res)
	if err == nil {
		t.Fatal("want an error from the cancelled context")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want it to wrap context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("returned after %v; the context did not cancel the request", elapsed)
	}
}

// A non-OK token response used to return nil, leaving an empty token that went
// out as "Authorization: Bearer ". Bad credentials looked like missing data.
func TestBadCredentialsAreReported(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer srv.Close()

	c := NewCustomClient(srv.URL, srv.URL+"/token", "test")
	c.UseAuth("id", "wrong")

	var res Response[[]map[string]any]
	err := StationType(context.Background(), c, DefaultRequest(), &res)
	if err == nil {
		t.Fatal("want an error for rejected credentials, got nil")
	}
}

// A non-2xx from the API is typed, so a caller can distinguish a 404 from a
// transport failure without matching on a string.
func TestStatusErrorIsTyped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewCustomClient(srv.URL, "", "test")
	var res Response[[]map[string]any]
	err := StationType(context.Background(), c, DefaultRequest(), &res)

	var se *StatusError
	if !errors.As(err, &se) {
		t.Fatalf("err = %v, want a *StatusError", err)
	}
	if se.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", se.StatusCode)
	}
}
