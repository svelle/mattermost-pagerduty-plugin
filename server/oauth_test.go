// Copyright (c) 2026-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/svelle/mattermost-pagerduty-plugin/server/store/kvstore"
)

// tokenState is a thread-safe holder for a user's stored OAuth token, used to
// observe what the refresh logic reads, writes, and deletes.
type tokenState struct {
	mu       sync.Mutex
	token    *kvstore.OAuthToken
	deleted  bool
	setCount int
}

// newStatefulStore returns a mockKVStore whose user-token methods are backed by a
// thread-safe in-memory token, plus the shared state for assertions.
func newStatefulStore(initial *kvstore.OAuthToken) (*mockKVStore, *tokenState) {
	st := &tokenState{token: initial}
	return &mockKVStore{
		getUserTokenFunc: func(_ string) (*kvstore.OAuthToken, error) {
			st.mu.Lock()
			defer st.mu.Unlock()
			if st.token == nil {
				return nil, nil
			}
			cp := *st.token
			return &cp, nil
		},
		setUserTokenFunc: func(_ string, token *kvstore.OAuthToken) error {
			st.mu.Lock()
			defer st.mu.Unlock()
			cp := *token
			st.token = &cp
			st.setCount++
			return nil
		},
		deleteUserTokenFunc: func(_ string) error {
			st.mu.Lock()
			defer st.mu.Unlock()
			st.token = nil
			st.deleted = true
			return nil
		},
	}, st
}

func newOAuthTestPlugin(tokenURL string, store kvstore.KVStore) *Plugin {
	p := newTestPlugin(newMockAPI())
	p.kvstore = store
	p.configuration = &configuration{
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
		APIBaseURL:        "https://api.pagerduty.com",
	}
	p.tokenURLOverride = tokenURL
	return p
}

func expiredToken() *kvstore.OAuthToken {
	return &kvstore.OAuthToken{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		TokenType:    "bearer",
		ExpiresAt:    time.Now().Add(-time.Minute),
	}
}

// TestGetPagerDutyClientForUser_SingleFlightRefresh verifies that many concurrent
// callers with an expired token trigger exactly one refresh (no stale-token replay).
func TestGetPagerDutyClientForUser_SingleFlightRefresh(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		// Widen the race window so concurrent callers pile up on the lock.
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"new-refresh","token_type":"bearer","expires_in":3600}`))
	}))
	defer server.Close()

	store, state := newStatefulStore(expiredToken())
	p := newOAuthTestPlugin(server.URL, store)

	const n = 25
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			client, err := p.getPagerDutyClientForUser("user-1")
			errs[idx] = err
			if err == nil {
				assert.NotNil(t, client)
			}
		}(i)
	}
	wg.Wait()

	for _, err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), atomic.LoadInt32(&hits), "refresh endpoint must be hit exactly once")

	state.mu.Lock()
	require.NotNil(t, state.token)
	assert.Equal(t, "new-access", state.token.AccessToken)
	assert.Equal(t, "new-refresh", state.token.RefreshToken)
	assert.Equal(t, 1, state.setCount, "token must be persisted exactly once")
	state.mu.Unlock()
}

// TestRefreshUserToken_InvalidGrantIsTerminal verifies invalid_grant clears the
// stored token and surfaces ErrTokenExpired so the user is prompted to reconnect.
func TestRefreshUserToken_InvalidGrantIsTerminal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid or unknown refresh token provided."}`))
	}))
	defer server.Close()

	store, state := newStatefulStore(expiredToken())
	p := newOAuthTestPlugin(server.URL, store)

	_, err := p.getPagerDutyClientForUser("user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenExpired), "invalid_grant must be terminal")
	assert.False(t, errors.Is(err, ErrTokenRefreshUnavailable))

	state.mu.Lock()
	assert.True(t, state.deleted, "token must be deleted on invalid_grant")
	assert.Nil(t, state.token)
	state.mu.Unlock()
}

// TestRefreshUserToken_ServerErrorIsTransient verifies a 5xx is retryable and does
// not discard the stored token or force a reconnect.
func TestRefreshUserToken_ServerErrorIsTransient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server_error"}`))
	}))
	defer server.Close()

	store, state := newStatefulStore(expiredToken())
	p := newOAuthTestPlugin(server.URL, store)

	_, err := p.getPagerDutyClientForUser("user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenRefreshUnavailable), "5xx must be transient")
	assert.False(t, errors.Is(err, ErrTokenExpired))

	state.mu.Lock()
	assert.False(t, state.deleted, "token must be preserved on transient error")
	require.NotNil(t, state.token)
	assert.Equal(t, "old-refresh", state.token.RefreshToken)
	state.mu.Unlock()
}

// TestRefreshUserToken_NetworkErrorIsTransient verifies a network failure reaching
// the token endpoint is treated as transient and preserves the token.
func TestRefreshUserToken_NetworkErrorIsTransient(t *testing.T) {
	// Point at a server that is immediately closed so the request fails to connect.
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	store, state := newStatefulStore(expiredToken())
	p := newOAuthTestPlugin(url, store)

	_, err := p.getPagerDutyClientForUser("user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTokenRefreshUnavailable))

	state.mu.Lock()
	assert.False(t, state.deleted)
	require.NotNil(t, state.token)
	state.mu.Unlock()
}

// TestRefreshUserToken_PreservesRefreshTokenWhenOmitted verifies that when the
// refresh response omits a refresh_token, the existing one is retained.
func TestRefreshUserToken_PreservesRefreshTokenWhenOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","token_type":"bearer","expires_in":3600}`))
	}))
	defer server.Close()

	initial := expiredToken()
	initial.RefreshToken = "keep-me"
	store, state := newStatefulStore(initial)
	p := newOAuthTestPlugin(server.URL, store)

	client, err := p.getPagerDutyClientForUser("user-1")
	require.NoError(t, err)
	assert.NotNil(t, client)

	state.mu.Lock()
	require.NotNil(t, state.token)
	assert.Equal(t, "new-access", state.token.AccessToken)
	assert.Equal(t, "keep-me", state.token.RefreshToken)
	state.mu.Unlock()
}

// TestGetPagerDutyClientForUser_NotConnected verifies the no-token path is unchanged.
func TestGetPagerDutyClientForUser_NotConnected(t *testing.T) {
	store, _ := newStatefulStore(nil)
	p := newOAuthTestPlugin("http://127.0.0.1:0", store)

	_, err := p.getPagerDutyClientForUser("user-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNotConnected))
}
