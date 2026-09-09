// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskExecutionDTOJSON(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	dto := TaskExecutionDTO{
		ID:           12345,
		TaskID:       "task-abc",
		TaskType:     "send_email",
		TaskName:     "Send Email",
		Status:       "success",
		Retryable:    true,
		MaxRetry:     3,
		RetryCount:   1,
		Log:          "Task finished successfully",
		ErrorMessage: "",
		Result:       `{"status":"ok"}`,
		StartedAt:    &now,
		FinishedAt:   &now,
		Duration:     500,
		Payload:      `{"to":"test@example.com"}`,
		TriggeredBy:  TaskTriggerManual,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("json.Marshal(dto) error: %v", err)
	}

	var parsed TaskExecutionDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.ID != dto.ID {
		t.Errorf("got ID = %d, want %d", parsed.ID, dto.ID)
	}
	if parsed.TriggeredBy != TaskTriggerManual {
		t.Errorf("got TriggeredBy = %q, want %q", parsed.TriggeredBy, TaskTriggerManual)
	}
	if parsed.TaskType != "send_email" {
		t.Errorf("got TaskType = %q, want %q", parsed.TaskType, "send_email")
	}
}

func TestTaskMetaDTOJSON(t *testing.T) {
	meta := TaskMetaDTO{
		Type:        "cleanup",
		AsynqTask:   "task:cleanup",
		Name:        "Clean Cache",
		DisplayName: "Clean Cache",
		Description: "Cleans expired cache",
		Category:    "maintenance",
		Params: []TaskParamDTO{
			{
				Name:     "dry_run",
				Label:    "Dry Run",
				Type:     "boolean",
				Required: false,
				Default:  false,
			},
		},
		MaxRetry:  1,
		Timeout:   10 * time.Minute,
		Queue:     "default",
		Retryable: false,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("json.Marshal(meta) error: %v", err)
	}

	var parsed TaskMetaDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Type != meta.Type || len(parsed.Params) != 1 {
		t.Errorf("parsed DTO mismatch: %+v", parsed)
	}
	if parsed.Params[0].Name != "dry_run" {
		t.Errorf("expected param name dry_run, got %s", parsed.Params[0].Name)
	}
}
