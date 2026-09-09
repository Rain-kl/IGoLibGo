// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import "testing"

func TestS3BackendKey(t *testing.T) {
	tests := []struct {
		name      string
		keyPrefix string
		inputKey  string
		expected  string
	}{
		{
			name:      "empty prefix",
			keyPrefix: "",
			inputKey:  "uploads/avatar.png",
			expected:  "uploads/avatar.png",
		},
		{
			name:      "empty prefix with leading slash",
			keyPrefix: "",
			inputKey:  "/uploads/avatar.png",
			expected:  "uploads/avatar.png",
		},
		{
			name:      "with prefix",
			keyPrefix: "my-bucket-prefix",
			inputKey:  "avatar.png",
			expected:  "my-bucket-prefix/avatar.png",
		},
		{
			name:      "already has prefix",
			keyPrefix: "my-bucket-prefix",
			inputKey:  "my-bucket-prefix/avatar.png",
			expected:  "my-bucket-prefix/avatar.png",
		},
		{
			name:      "already has prefix with leading slash",
			keyPrefix: "my-bucket-prefix",
			inputKey:  "/my-bucket-prefix/avatar.png",
			expected:  "my-bucket-prefix/avatar.png",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &s3Backend{keyPrefix: tc.keyPrefix}
			actual := b.key(tc.inputKey)
			if actual != tc.expected {
				t.Errorf("s3Backend.key(%q) = %q, want %q", tc.inputKey, actual, tc.expected)
			}
		})
	}
}
