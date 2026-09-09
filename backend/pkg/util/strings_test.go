// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func TestDerefString(t *testing.T) {
	if got := DerefString(nil); got != "" {
		t.Errorf("DerefString(nil) = %q, want empty string", got)
	}
	s := "hello"
	if got := DerefString(&s); got != "hello" {
		t.Errorf("DerefString(&s) = %q, want %q", got, "hello")
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"invalid-email", "invalid-email"},
		{"a@b.com", "**@b.com"},
		{"ab@b.com", "**@b.com"},
		{"user@example.com", "us***r@example.com"},
		{"antigravity@wavelet.dev", "an***y@wavelet.dev"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := MaskEmail(tc.input)
			if got != tc.want {
				t.Errorf("MaskEmail(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTrimStringFields(t *testing.T) {
	s1 := "  hello  "
	s2 := "\tworld\n"
	var s3 *string

	TrimStringFields(&s1, &s2, s3)

	if s1 != "hello" {
		t.Errorf("expected s1 = %q, got %q", "hello", s1)
	}
	if s2 != "world" {
		t.Errorf("expected s2 = %q, got %q", "world", s2)
	}
}

func TestGenerateUniqueIDSimple(t *testing.T) {
	id1 := GenerateUniqueIDSimple()
	id2 := GenerateUniqueIDSimple()

	if len(id1) != 64 {
		t.Errorf("expected length 64, got %d (%s)", len(id1), id1)
	}
	if len(id2) != 64 {
		t.Errorf("expected length 64, got %d (%s)", len(id2), id2)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, got identical: %s", id1)
	}
}

func TestInterface2String(t *testing.T) {
	if got := Interface2String("test"); got != "test" {
		t.Errorf("got %q, want %q", got, "test")
	}
	if got := Interface2String(42); got != "42" {
		t.Errorf("got %q, want %q", got, "42")
	}
	if got := Interface2String(true); got != "Not Implemented" {
		t.Errorf("got %q, want %q", got, "Not Implemented")
	}
}
