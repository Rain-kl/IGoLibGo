// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"testing"
	"time"
)

func TestUnique(t *testing.T) {
	if got := Unique[int](nil); got != nil {
		t.Errorf("Unique(nil) = %v, want nil", got)
	}

	input := []int{1, 2, 2, 3, 1, 4, 3, 5}
	want := []int{1, 2, 3, 4, 5}
	got := Unique(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Unique(%v) = %v, want %v", input, got, want)
	}
}

func TestUniqueAndCleanStringSlice(t *testing.T) {
	if got := UniqueAndCleanStringSlice(nil); got != nil {
		t.Errorf("UniqueAndCleanStringSlice(nil) = %v, want nil", got)
	}

	emptyInput := []string{"  ", "", "\t"}
	if got := UniqueAndCleanStringSlice(emptyInput); got != nil {
		t.Errorf("UniqueAndCleanStringSlice(emptyInput) = %v, want nil", got)
	}

	input := []string{"  a ", "b", "a", " c ", "b", ""}
	want := []string{"a", "b", "c"}
	got := UniqueAndCleanStringSlice(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqueAndCleanStringSlice(%v) = %v, want %v", input, got, want)
	}
}

type testRecord struct {
	id uint
	tm time.Time
}

func (r testRecord) GetID() uint        { return r.id }
func (r testRecord) GetTime() time.Time { return r.tm }

func TestSortAndLimitRecords(t *testing.T) {
	now := time.Now()
	t1 := now.Add(-2 * time.Hour)
	t2 := now.Add(-1 * time.Hour)
	t3 := now

	records := []testRecord{
		{id: 1, tm: t1},
		{id: 2, tm: t3},
		{id: 3, tm: t2},
		{id: 4, tm: t3}, // same timestamp as id 2, higher id
	}

	// Should sort descending by time, then descending by ID
	sorted := SortAndLimitRecords(records, 3)
	if len(sorted) != 3 {
		t.Fatalf("expected 3 records, got %d", len(sorted))
	}
	if sorted[0].id != 4 {
		t.Errorf("expected sorted[0].id = 4, got %d", sorted[0].id)
	}
	if sorted[1].id != 2 {
		t.Errorf("expected sorted[1].id = 2, got %d", sorted[1].id)
	}
	if sorted[2].id != 3 {
		t.Errorf("expected sorted[2].id = 3, got %d", sorted[2].id)
	}
}

func TestMap(t *testing.T) {
	if got := Map[int, string](nil, func(i int) string { return "" }); got != nil {
		t.Errorf("Map(nil) = %v, want nil", got)
	}

	input := []int{1, 2, 3}
	got := Map(input, func(i int) int { return i * 2 })
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Map(%v) = %v, want %v", input, got, want)
	}
}

func TestFilter(t *testing.T) {
	if got := Filter[int](nil, func(i int) bool { return true }); got != nil {
		t.Errorf("Filter(nil) = %v, want nil", got)
	}

	input := []int{1, 2, 3, 4, 5, 6}
	got := Filter(input, func(i int) bool { return i%2 == 0 })
	want := []int{2, 4, 6}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Filter(%v) = %v, want %v", input, got, want)
	}
}
