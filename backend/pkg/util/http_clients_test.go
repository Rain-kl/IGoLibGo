// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsLocalhost(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://localhost:8080/api", true},
		{"https://127.0.0.1:3000", true},
		{"http://[::1]:8080", true},
		{"https://example.com/api", false},
		{"invalid url ://", false},
	}

	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			if got := IsLocalhost(tc.url); got != tc.expected {
				t.Errorf("IsLocalhost(%q) = %v, want %v", tc.url, got, tc.expected)
			}
		})
	}
}

func TestRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "test-token" {
			http.Error(w, "missing cookie", http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Custom") != "CustomValue" {
			http.Error(w, "missing header", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	ctx := context.Background()
	headers := map[string]string{"X-Custom": "CustomValue"}
	cookies := map[string]string{"session": "test-token"}

	resp, err := Request(ctx, http.MethodGet, server.URL, nil, headers, cookies)
	if err != nil {
		t.Fatalf("Request() returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("expected body %q, got %q", "ok", string(body))
	}
}
