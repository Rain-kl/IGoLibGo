// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRateAndResultJSON(t *testing.T) {
	rate := Rate{
		Limit:  100,
		Period: time.Minute,
	}

	data, err := json.Marshal(rate)
	if err != nil {
		t.Fatalf("json.Marshal(rate) error: %v", err)
	}

	var parsedRate Rate
	if err := json.Unmarshal(data, &parsedRate); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsedRate.Limit != rate.Limit || parsedRate.Period != rate.Period {
		t.Errorf("rate mismatch: got %+v, want %+v", parsedRate, rate)
	}

	result := RateLimitResult{
		Allowed:    true,
		Remaining:  99,
		ResetAfter: 50 * time.Second,
		RetryAfter: 0,
	}

	resData, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal(result) error: %v", err)
	}

	var parsedResult RateLimitResult
	if err := json.Unmarshal(resData, &parsedResult); err != nil {
		t.Fatalf("json.Unmarshal(resData) error: %v", err)
	}

	if parsedResult.Allowed != result.Allowed || parsedResult.Remaining != result.Remaining {
		t.Errorf("result mismatch: got %+v, want %+v", parsedResult, result)
	}
}
