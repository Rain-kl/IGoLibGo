// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
)

func TestDomainEventsSerialization(t *testing.T) {
	t.Run("TaskCompletedEvent", func(t *testing.T) {
		evt := TaskCompletedEvent{
			TaskID:    "t-123",
			TaskName:  "Test Task",
			TaskType:  "test",
			Status:    "success",
			Duration:  150,
			ResultMsg: "done",
		}
		data, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var parsed TaskCompletedEvent
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if parsed.TaskID != evt.TaskID || parsed.ResultMsg != evt.ResultMsg {
			t.Errorf("mismatch: got %+v, want %+v", parsed, evt)
		}
	})

	t.Run("UploadCreatedEvent", func(t *testing.T) {
		evt := UploadCreatedEvent{
			UploadID: 456,
			UserID:   123,
			FileName: "photo.jpg",
			FileSize: 1024,
			MimeType: "image/jpeg",
		}
		data, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var parsed UploadCreatedEvent
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if parsed.UploadID != evt.UploadID || parsed.FileName != evt.FileName {
			t.Errorf("mismatch: got %+v, want %+v", parsed, evt)
		}
	})

	t.Run("NotificationSentEvent", func(t *testing.T) {
		evt := NotificationSentEvent{
			UserID:  123,
			Channel: "telegram",
			Title:   "Notice",
			Success: true,
		}
		data, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var parsed NotificationSentEvent
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if parsed.Channel != evt.Channel || !parsed.Success {
			t.Errorf("mismatch: got %+v, want %+v", parsed, evt)
		}
	})

	t.Run("ConfigChangedEvent", func(t *testing.T) {
		evt := ConfigChangedEvent{
			Key:    "site_name",
			OldVal: "Wavelet Old",
			NewVal: "Wavelet New",
		}
		data, err := json.Marshal(evt)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}
		var parsed ConfigChangedEvent
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}
		if parsed.Key != evt.Key {
			t.Errorf("mismatch: got %+v, want %+v", parsed, evt)
		}
	})
}
