// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		left     string
		right    string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"v1.0.0", "1.0.0", 0},
		{"1.0.0", "1.0.1", -1},
		{"1.1.0", "1.0.1", 1},
		{"2.0.0", "1.9.9", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.beta", "1.0.0-beta", -1},
		{"dev", "1.0.0", -1},
		{"dev", "dev", 0},
		{"1.0.0", "dev", 0},
		{"1.0.0-1-g1234567", "1.0.0-2-g1234567", -1},
		{"1.0.0-2-g1234567", "1.0.0-1-g1234567", 1},
		{"1.0.0-1-g1234567", "1.0.0-1-g1234567", 0},
	}

	for _, tc := range tests {
		t.Run(tc.left+"_vs_"+tc.right, func(t *testing.T) {
			got := CompareVersions(tc.left, tc.right)
			if got != tc.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tc.left, tc.right, got, tc.expected)
			}
		})
	}
}

func TestParseVersionInfo(t *testing.T) {
	info := ParseVersionInfo("v1.2.3-alpha.4")
	if !info.Valid {
		t.Fatal("expected Valid to be true")
	}
	if len(info.Numbers) != 3 || info.Numbers[0] != 1 || info.Numbers[1] != 2 || info.Numbers[2] != 3 {
		t.Errorf("unexpected Numbers: %v", info.Numbers)
	}
	if len(info.Prerelease) != 2 || info.Prerelease[0] != "alpha" || info.Prerelease[1] != "4" {
		t.Errorf("unexpected Prerelease: %v", info.Prerelease)
	}

	devInfo := ParseVersionInfo("dev")
	if !devInfo.IsDev {
		t.Error("expected IsDev to be true")
	}
}
