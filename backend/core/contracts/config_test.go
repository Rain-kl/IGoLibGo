// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSystemConfigDTOJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	cfg := SystemConfigDTO{
		Key:         "site_name",
		Value:       "Wavelet",
		Type:        "system",
		Visibility:  1,
		Description: "Site Name",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(cfg) error: %v", err)
	}

	var parsed SystemConfigDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Key != cfg.Key || parsed.Value != cfg.Value || parsed.Visibility != 1 {
		t.Errorf("mismatch: got %+v, want %+v", parsed, cfg)
	}
}
