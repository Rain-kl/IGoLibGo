// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"reflect"
	"testing"
)

func TestUploadMetadataDTOValueAndScan(t *testing.T) {
	meta := UploadMetadataDTO{
		Width:        1920,
		Height:       1080,
		Duration:     12.5,
		OriginalMime: "video/mp4",
		Bucket:       "videos",
		Extra:        map[string]any{"codec": "h264"},
	}

	val, err := meta.Value()
	if err != nil {
		t.Fatalf("UploadMetadataDTO.Value() error: %v", err)
	}

	bytesVal, ok := val.([]byte)
	if !ok {
		t.Fatalf("expected []byte, got %T", val)
	}

	var scanned UploadMetadataDTO
	if err := scanned.Scan(bytesVal); err != nil {
		t.Fatalf("UploadMetadataDTO.Scan(bytes) error: %v", err)
	}
	if scanned.Width != meta.Width || scanned.OriginalMime != meta.OriginalMime {
		t.Errorf("got %+v, want %+v", scanned, meta)
	}

	// Scan string
	var scannedStr UploadMetadataDTO
	if err := scannedStr.Scan(string(bytesVal)); err != nil {
		t.Fatalf("UploadMetadataDTO.Scan(string) error: %v", err)
	}
	if scannedStr.Width != meta.Width {
		t.Errorf("got width %d, want %d", scannedStr.Width, meta.Width)
	}

	// Scan nil
	var scannedNil UploadMetadataDTO
	if err := scannedNil.Scan(nil); err != nil {
		t.Fatalf("UploadMetadataDTO.Scan(nil) error: %v", err)
	}
	if !reflect.DeepEqual(scannedNil, UploadMetadataDTO{}) {
		t.Errorf("expected empty DTO on nil scan, got %+v", scannedNil)
	}

	// Scan invalid type
	if err := scanned.Scan(12345); err == nil {
		t.Errorf("expected error scanning int, got nil")
	}
}
