// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import "testing"

func TestWebDAVBackendKey(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		inputKey string
		expected string
	}{
		{
			name:     "empty base path",
			basePath: "",
			inputKey: "uploads/test.txt",
			expected: "/uploads/test.txt",
		},
		{
			name:     "empty base path with leading slash",
			basePath: "",
			inputKey: "/uploads/test.txt",
			expected: "/uploads/test.txt",
		},
		{
			name:     "with base path",
			basePath: "remote/dav",
			inputKey: "uploads/test.txt",
			expected: "/remote/dav/uploads/test.txt",
		},
		{
			name:     "with base path and leading slash on input",
			basePath: "remote/dav",
			inputKey: "/uploads/test.txt",
			expected: "/remote/dav/uploads/test.txt",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &webDAVBackend{basePath: tc.basePath}
			actual := b.key(tc.inputKey)
			if actual != tc.expected {
				t.Errorf("webDAVBackend.key(%q) = %q, want %q", tc.inputKey, actual, tc.expected)
			}
		})
	}
}
