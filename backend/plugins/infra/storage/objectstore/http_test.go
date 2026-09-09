// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHTTPObject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/files/hello.txt":
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte("hello world"))
		case "/files/no-mime.bin":
			// No Content-Type header sent
			_, _ = w.Write([]byte("binary content"))
		case "/files/not-found":
			http.NotFound(w, r)
		default:
			http.Error(w, "unexpected", http.StatusBadRequest)
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	t.Run("success with mime", func(t *testing.T) {
		obj, err := getHTTPObject(ctx, ts.URL, "files/hello.txt")
		if err != nil {
			t.Fatalf("getHTTPObject() unexpected error: %v", err)
		}
		defer obj.Body.Close()

		data, err := io.ReadAll(obj.Body)
		if err != nil {
			t.Fatalf("ReadAll() error: %v", err)
		}
		if string(data) != "hello world" {
			t.Errorf("got %q, want %q", string(data), "hello world")
		}
		if obj.ContentType != "text/plain; charset=utf-8" {
			t.Errorf("got content type %q, want %q", obj.ContentType, "text/plain; charset=utf-8")
		}
	})

	t.Run("default content type fallback", func(t *testing.T) {
		obj, err := getHTTPObject(ctx, ts.URL, "files/no-mime.bin")
		if err != nil {
			t.Fatalf("getHTTPObject() unexpected error: %v", err)
		}
		defer obj.Body.Close()

		if obj.ContentType == "" {
			t.Errorf("expected non-empty default content type")
		}
	})

	t.Run("non-200 error", func(t *testing.T) {
		_, err := getHTTPObject(ctx, ts.URL, "files/not-found")
		if err == nil {
			t.Fatal("expected error for 404 response, got nil")
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		_, err := getHTTPObject(ctx, "http://[invalid-host", "key")
		if err == nil {
			t.Fatal("expected error for invalid url, got nil")
		}
	})
}
