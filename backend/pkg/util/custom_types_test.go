// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"testing"
)

func TestStringArrayValueAndScan(t *testing.T) {
	sa := StringArray{"alpha", "beta", "gamma"}

	val, err := sa.Value()
	if err != nil {
		t.Fatalf("StringArray.Value() returned error: %v", err)
	}

	bytesVal, ok := val.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", val)
	}

	var scanned StringArray
	if err := scanned.Scan(bytesVal); err != nil {
		t.Fatalf("StringArray.Scan() returned error: %v", err)
	}

	if !reflect.DeepEqual(sa, scanned) {
		t.Errorf("got %v, want %v", scanned, sa)
	}

	// Test invalid scan type
	if err := scanned.Scan(12345); err == nil {
		t.Errorf("expected error scanning int, got nil")
	}
}
