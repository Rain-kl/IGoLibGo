// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package entity

import (
	"Wavelet/igo-lib/plugins/igo/consts"
	"time"
)

// TaskRun is the latest coordinator snapshot for one kind per user.
type TaskRun struct {
	ID            uint64     `gorm:"primaryKey" json:"id,string"`
	UserID        uint64     `gorm:"uniqueIndex:uq_igo_task_runs_user_kind,priority:1;not null" json:"user_id,string"`
	Kind          string     `gorm:"uniqueIndex:uq_igo_task_runs_user_kind,priority:2;size:32;not null" json:"kind"`
	State         string     `gorm:"size:32;not null" json:"state"`
	Title         string     `gorm:"size:128;not null;default:''" json:"title"`
	Message       string     `gorm:"type:text;not null;default:''" json:"message"`
	PlanJSON      string     `gorm:"type:text;not null;default:'{}'" json:"plan_json"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	LastRequestAt *time.Time `json:"last_request_at,omitempty"`
	PollCount     int        `gorm:"not null;default:0" json:"poll_count"`
	RequestCount  int        `gorm:"not null;default:0" json:"request_count"`
	Reason        string     `gorm:"size:64;not null;default:''" json:"reason"`
}

// TableName returns the plugin-prefixed table name.
func (TaskRun) TableName() string { return consts.TableTaskRuns }

// TaskLaunchHistory is a recent grab/global-leak launch record.
type TaskLaunchHistory struct {
	ID          uint64    `gorm:"primaryKey" json:"id,string"`
	UserID      uint64    `gorm:"uniqueIndex:uq_igo_task_history_record,priority:1;uniqueIndex:uq_igo_task_history_fingerprint,priority:1;index:idx_igo_task_history_user_kind,priority:1;not null" json:"user_id,string"`
	RecordID    string    `gorm:"uniqueIndex:uq_igo_task_history_record,priority:2;size:64;not null" json:"record_id"`
	Kind        string    `gorm:"uniqueIndex:uq_igo_task_history_fingerprint,priority:2;index:idx_igo_task_history_user_kind,priority:2;size:32;not null" json:"kind"`
	Fingerprint string    `gorm:"uniqueIndex:uq_igo_task_history_fingerprint,priority:3;size:128;not null" json:"fingerprint"`
	RecordedAt  time.Time `gorm:"not null" json:"recorded_at"`
	PayloadJSON string    `gorm:"type:text;not null" json:"payload_json"`
}

// TableName returns the plugin-prefixed table name.
func (TaskLaunchHistory) TableName() string { return consts.TableTaskLaunchHistory }
