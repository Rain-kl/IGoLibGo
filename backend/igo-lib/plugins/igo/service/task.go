// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/model/entity"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

const (
	stateIdle    = "idle"
	stateRunning = "running"
	stateSuccess = "succeeded"
	stateFailed  = "failed"

	maxOccupyRetries        = 3
	schedulePreWaitInterval = 30 * time.Second
	defaultPollInterval     = 10 * time.Second
	occupyBaseIntervalSec   = 10
	occupyJitterRangeSec    = 11
	aggressiveDelayMS       = 1000
	randomMinDelayMS        = 4000
	randomMaxDelayMS        = 8000
	defaultDelayMS          = 5000
	dayDuration             = 24 * time.Hour
)

// ListTasks returns coordinator snapshots.
func (s *Service) ListTasks(ctx context.Context, userID uint64) (*do.TaskListResponse, error) {
	rows, err := dao.ListTaskRuns(ctx, userID)
	if err != nil {
		return nil, err
	}
	byKind := map[string]entity.TaskRun{}
	for _, r := range rows {
		byKind[r.Kind] = r
	}
	kinds := []string{consts.TaskKindGrab, consts.TaskKindOccupy, consts.TaskKindGlobalLeak, consts.TaskKindTomorrow}
	out := make([]do.CoordinatorStatus, 0, len(kinds))
	for _, kind := range kinds {
		if r, ok := byKind[kind]; ok {
			out = append(out, toStatus(r))
			continue
		}
		out = append(out, do.CoordinatorStatus{Kind: kind, State: stateIdle, Title: kindTitle(kind), Message: "未运行"})
	}
	return &do.TaskListResponse{Tasks: out}, nil
}

// GetTaskStatus returns the latest status of a coordinator.
func (s *Service) GetTaskStatus(ctx context.Context, userID uint64, kind string) (*do.CoordinatorStatus, error) {
	run, err := dao.GetTaskRun(ctx, userID, kind)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return &do.CoordinatorStatus{
			Kind:    kind,
			State:   stateIdle,
			Title:   kindTitle(kind),
			Message: "未运行",
		}, nil
	}
	st := toStatus(*run)
	return &st, nil
}

// ListTaskRecords returns recent launch history.
func (s *Service) ListTaskRecords(ctx context.Context, userID uint64, kindFilter ...string) ([]do.TaskLaunchRecord, error) {
	var out []do.TaskLaunchRecord
	targetKinds := []string{consts.TaskKindGrab, consts.TaskKindGlobalLeak}
	if len(kindFilter) > 0 && kindFilter[0] != "" {
		targetKinds = []string{kindFilter[0]}
	}
	for _, kind := range targetKinds {
		rows, err := dao.ListTaskLaunchHistory(ctx, userID, kind, consts.MaxTaskLaunchHistory)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			rec := do.TaskLaunchRecord{
				RecordID:   r.RecordID,
				Kind:       r.Kind,
				RecordedAt: r.RecordedAt.UTC().Format(time.RFC3339),
			}
			_ = json.Unmarshal([]byte(r.PayloadJSON), &rec)
			rec.RecordID = r.RecordID
			rec.Kind = r.Kind
			rec.RecordedAt = r.RecordedAt.UTC().Format(time.RFC3339)
			out = append(out, rec)
		}
	}
	return out, nil
}

// StartGrab starts the grab coordinator.
func (s *Service) StartGrab(ctx context.Context, userID uint64, req do.GrabStartRequest) error {
	return s.startKind(ctx, userID, consts.TaskKindGrab, "抢座任务已启动", req)
}

// StartOccupy starts the occupy coordinator.
func (s *Service) StartOccupy(ctx context.Context, userID uint64, req do.OccupyStartRequest) error {
	return s.startKind(ctx, userID, consts.TaskKindOccupy, "占座任务已启动", req)
}

// StartGlobalLeak starts the global-leak coordinator.
func (s *Service) StartGlobalLeak(ctx context.Context, userID uint64, req do.GlobalLeakStartRequest) error {
	return s.startKind(ctx, userID, consts.TaskKindGlobalLeak, "全域捡漏已启动", req)
}

// StartTomorrow starts the tomorrow-reservation coordinator.
func (s *Service) StartTomorrow(ctx context.Context, userID uint64, req do.TomorrowStartRequest) error {
	return s.startKind(ctx, userID, consts.TaskKindTomorrow, "明日预约已启动", req)
}

// RunTomorrowNow executes one tomorrow-reservation attempt immediately.
func (s *Service) RunTomorrowNow(ctx context.Context, userID uint64, req do.TomorrowStartRequest) error {
	req.ExecuteImmediately = true
	if err := s.startKind(ctx, userID, consts.TaskKindTomorrow, "明日预约立即执行", req); err != nil {
		return err
	}
	return s.runTick(ctx, userID, consts.TaskKindTomorrow)
}

// CancelTask stops a coordinator by kind.
func (s *Service) CancelTask(ctx context.Context, userID uint64, kind string) error {
	now := time.Now().UTC()
	return dao.UpsertTaskRun(ctx, &entity.TaskRun{
		UserID:        userID,
		Kind:          kind,
		State:         stateIdle,
		Title:         kindTitle(kind),
		Message:       "已停止",
		LastUpdatedAt: &now,
	})
}

// GetGlobalLeakBlacklist returns blacklisted seats keyed by library id.
func (s *Service) GetGlobalLeakBlacklist(ctx context.Context, userID uint64) (*do.GlobalLeakBlacklistResponse, error) {
	rows, err := dao.ListGlobalLeakBlacklist(ctx, userID, nil)
	if err != nil {
		return nil, err
	}
	items := map[int][]do.SeatRef{}
	for _, r := range rows {
		items[r.LibraryID] = append(items[r.LibraryID], do.SeatRef{SeatKey: r.SeatKey, SeatName: r.SeatName})
	}
	return &do.GlobalLeakBlacklistResponse{Items: items}, nil
}

// SaveGlobalLeakBlacklist replaces submitted venue blacklists.
func (s *Service) SaveGlobalLeakBlacklist(ctx context.Context, userID uint64, req do.SaveGlobalLeakBlacklistRequest) error {
	var libIDs []int
	var seats []entity.GlobalLeakBlacklistSeat
	for id, list := range req.Items {
		libIDs = append(libIDs, id)
		for _, seat := range list {
			seats = append(seats, entity.GlobalLeakBlacklistSeat{LibraryID: id, SeatKey: seat.SeatKey, SeatName: seat.SeatName})
		}
	}
	return dao.ReplaceGlobalLeakBlacklistForLibraries(ctx, userID, libIDs, seats)
}

// GetGlobalLeakSelectedLibraries returns scan-priority venues.
func (s *Service) GetGlobalLeakSelectedLibraries(ctx context.Context, userID uint64) ([]do.GlobalLeakLibraryTarget, error) {
	rows, err := dao.ListGlobalLeakTargets(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]do.GlobalLeakLibraryTarget, 0, len(rows))
	for _, r := range rows {
		out = append(out, do.GlobalLeakLibraryTarget{LibraryID: r.LibraryID, LibraryName: r.LibraryName, Floor: r.Floor})
	}
	return out, nil
}

// SaveGlobalLeakSelectedLibraries persists scan-priority venues.
func (s *Service) SaveGlobalLeakSelectedLibraries(ctx context.Context, userID uint64, req do.SaveGlobalLeakSelectedLibrariesRequest) error {
	items := make([]entity.GlobalLeakTarget, 0, len(req.Libraries))
	for i, lib := range req.Libraries {
		items = append(items, entity.GlobalLeakTarget{
			LibraryID:    lib.LibraryID,
			LibraryName:  lib.LibraryName,
			Floor:        lib.Floor,
			ScanPriority: i,
		})
	}
	return dao.ReplaceGlobalLeakTargets(ctx, userID, items)
}

func (s *Service) startKind(ctx context.Context, userID uint64, kind, message string, plan any) error {
	if existing, err := dao.GetTaskRun(ctx, userID, kind); err != nil {
		return err
	} else if existing != nil && existing.State == stateRunning {
		return consts.NewError(http.StatusConflict, consts.CodeConflict, kindTitle(kind)+"任务已在运行")
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if err := dao.UpsertTaskRun(ctx, &entity.TaskRun{
		UserID:        userID,
		Kind:          kind,
		State:         stateRunning,
		Title:         kindTitle(kind),
		Message:       message,
		PlanJSON:      string(raw),
		StartedAt:     &now,
		LastUpdatedAt: &now,
	}); err != nil {
		return err
	}
	s.recordLaunch(ctx, userID, kind, plan, raw)
	s.dispatchTick(ctx, userID, kind)
	return nil
}

func (s *Service) recordLaunch(ctx context.Context, userID uint64, kind string, _ any, raw []byte) {
	if kind != consts.TaskKindGrab && kind != consts.TaskKindGlobalLeak {
		return
	}
	sum := sha256.Sum256(raw)
	_ = dao.UpsertTaskLaunchHistory(ctx, &entity.TaskLaunchHistory{
		UserID:      userID,
		RecordID:    hex.EncodeToString(sum[:8]),
		Kind:        kind,
		Fingerprint: hex.EncodeToString(sum[:]),
		RecordedAt:  time.Now().UTC(),
		PayloadJSON: string(raw),
	})
}

func (s *Service) dispatchTick(ctx context.Context, userID uint64, kind string) {
	if s.tasks == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{"user_id": userID, "kind": kind})
	_, _ = s.tasks.Dispatch(ctx, consts.TaskTypeTick, payload, contracts.TaskTriggerSystem)
}

// HandleTick is the Asynq/inproc worker for coordinator polling.
func (s *Service) HandleTick(ctx context.Context, payload []byte) error {
	var p struct {
		UserID uint64 `json:"user_id"`
		Kind   string `json:"kind"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	return s.runTick(ctx, p.UserID, p.Kind)
}

func (s *Service) runTick(ctx context.Context, userID uint64, kind string) error {
	run, err := dao.GetTaskRun(ctx, userID, kind)
	if err != nil || run == nil || run.State != stateRunning {
		return err
	}
	cookie, _, err := s.cookie(ctx, userID)
	if err != nil {
		return s.failRun(ctx, run, err.Error())
	}
	s.maybeCookieAlert(ctx, userID, cookie)
	tpl, err := s.templates(ctx, userID)
	if err != nil {
		return s.failRun(ctx, run, err.Error())
	}

	ok, msg, err := s.tickOnce(ctx, userID, kind, cookie, tpl, run.PlanJSON)
	now := time.Now().UTC()
	run.PollCount++
	run.RequestCount++
	run.LastRequestAt = &now
	run.LastUpdatedAt = &now
	if err != nil {
		if isAuthFailure(err.Error()) {
			return s.failRun(ctx, run, err.Error())
		}
		run.Message = err.Error()
		if err := dao.UpsertTaskRun(ctx, run); err != nil {
			return err
		}
		if err := s.waitTick(ctx, kind, run.PlanJSON); err != nil {
			return err
		}
		s.dispatchTick(ctx, userID, kind)
		return nil
	}
	if ok {
		run.State = stateSuccess
		run.Message = msg
		run.Reason = kind + "_succeeded"
		_ = dao.UpsertTaskRun(ctx, run)
		s.bumpSuccess(ctx, userID)
		return nil
	}
	run.Message = msg
	if err := dao.UpsertTaskRun(ctx, run); err != nil {
		return err
	}
	if err := s.waitTick(ctx, kind, run.PlanJSON); err != nil {
		return err
	}
	s.dispatchTick(ctx, userID, kind)
	return nil
}

func (s *Service) tickOnce(ctx context.Context, userID uint64, kind, cookie string, tpl do.ProtocolTemplatesResponse, planJSON string) (bool, string, error) {
	switch kind {
	case consts.TaskKindGrab:
		var plan do.GrabStartRequest
		if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
			return false, "", err
		}
		return s.tickGrab(ctx, userID, tpl, cookie, plan)
	case consts.TaskKindOccupy:
		var plan do.OccupyStartRequest
		if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
			return false, "", err
		}
		return s.tickOccupy(ctx, userID, tpl, cookie, plan)
	case consts.TaskKindGlobalLeak:
		var plan do.GlobalLeakStartRequest
		if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
			return false, "", err
		}
		return s.tickLeak(ctx, userID, tpl, cookie, plan)
	case consts.TaskKindTomorrow:
		var plan do.TomorrowStartRequest
		if err := json.Unmarshal([]byte(planJSON), &plan); err != nil {
			return false, "", err
		}
		return s.tickTomorrow(ctx, userID, tpl, cookie, plan)
	default:
		return false, "", fmt.Errorf("unknown kind %s", kind)
	}
}

func (s *Service) tickGrab(ctx context.Context, userID uint64, tpl do.ProtocolTemplatesResponse, cookie string, plan do.GrabStartRequest) (bool, string, error) {
	if wait, msg := scheduledWait(plan.ScheduledStart); wait {
		return false, msg, nil
	}
	if plan.ReservationStrategy == "reserve_directly" {
		for _, seat := range plan.Seats {
			ok, err := s.api(ctx, userID).ReserveSeat(ctx, tpl, cookie, plan.LibraryID, seat.SeatKey)
			if err != nil {
				return false, "", err
			}
			if ok {
				msg := seat.SeatName + " 预约成功"
				s.notifySuccess(ctx, userID, consts.TaskKindGrab, plan.LibraryName, seat.SeatName, msg)
				return true, msg, nil
			}
		}
		return false, "直接预约未命中，继续", nil
	}
	layout, err := s.api(ctx, userID).GetLayout(ctx, tpl, cookie, plan.LibraryID)
	if err != nil {
		return false, "", err
	}
	wanted := map[string]do.SeatRef{}
	for _, seat := range plan.Seats {
		wanted[seat.SeatKey] = seat
	}
	for _, snap := range layout.Seats {
		if snap.IsOccupied {
			continue
		}
		if target, ok := wanted[snap.SeatKey]; ok {
			ok, err := s.api(ctx, userID).ReserveSeat(ctx, tpl, cookie, plan.LibraryID, snap.SeatKey)
			if err != nil {
				return false, "", err
			}
			if ok {
				name := target.SeatName
				if name == "" {
					name = snap.SeatName
				}
				msg := name + " 预约成功"
				s.notifySuccess(ctx, userID, consts.TaskKindGrab, plan.LibraryName, name, msg)
				return true, msg, nil
			}
		}
	}
	return false, "目标座位暂不可用，继续监控", nil
}

func (s *Service) tickOccupy(ctx context.Context, userID uint64, tpl do.ProtocolTemplatesResponse, cookie string, plan do.OccupyStartRequest) (bool, string, error) {
	info, err := s.api(ctx, userID).GetReservation(ctx, tpl, cookie)
	if err != nil {
		return false, "", err
	}
	if info == nil || !info.HasReservation {
		return false, "", fmt.Errorf("当前没有可续占的预约")
	}
	exp, err := time.Parse(time.RFC3339, info.ExpirationTime)
	if err != nil {
		return false, "", fmt.Errorf("预约到期时间无效")
	}
	if time.Until(exp) > 60*time.Second {
		return false, "预约仍有效，等待重预约窗口", nil
	}
	ok, err := s.api(ctx, userID).CancelReservation(ctx, tpl, cookie, info.ReservationToken)
	if err != nil {
		return false, "", err
	}
	if !ok {
		return false, "", fmt.Errorf("取消预约失败")
	}
	pause := time.Duration(plan.ReReserveDelaySeconds) * time.Second
	if pause > 0 {
		if err := sleepCtx(ctx, pause); err != nil {
			return false, "", err
		}
	}
	for attempt := 1; attempt <= maxOccupyRetries; attempt++ {
		ok, err = s.api(ctx, userID).ReserveSeat(ctx, tpl, cookie, info.LibraryID, info.SeatKey)
		if err != nil {
			return false, "", err
		}
		if ok {
			msg := info.SeatName + " 重新预约成功"
			s.notifySuccess(ctx, userID, consts.TaskKindOccupy, info.LibraryName, info.SeatName, msg)
			return false, msg, nil
		}
		if attempt < maxOccupyRetries {
			if err := sleepCtx(ctx, time.Second); err != nil {
				return false, "", err
			}
		}
	}
	return false, "", fmt.Errorf("重新预约失败，已达到重试上限")
}

func (s *Service) tickLeak(ctx context.Context, userID uint64, tpl do.ProtocolTemplatesResponse, cookie string, plan do.GlobalLeakStartRequest) (bool, string, error) {
	blocked := map[string]struct{}{}
	if bl, err := dao.ListGlobalLeakBlacklist(ctx, userID, nil); err == nil {
		for _, row := range bl {
			blocked[fmt.Sprintf("%d:%s", row.LibraryID, row.SeatKey)] = struct{}{}
		}
	}
	for _, lib := range plan.Libraries {
		layout, err := s.api(ctx, userID).GetLayout(ctx, tpl, cookie, lib.LibraryID)
		if err != nil {
			continue
		}
		for _, seat := range layout.Seats {
			if seat.IsOccupied {
				continue
			}
			if _, skip := blocked[fmt.Sprintf("%d:%s", lib.LibraryID, seat.SeatKey)]; skip {
				continue
			}
			ok, err := s.api(ctx, userID).ReserveSeat(ctx, tpl, cookie, lib.LibraryID, seat.SeatKey)
			if err != nil {
				return false, "", err
			}
			if ok {
				msg := lib.LibraryName + " " + seat.SeatName + " 捡漏成功"
				s.notifySuccess(ctx, userID, consts.TaskKindGlobalLeak, lib.LibraryName, seat.SeatName, msg)
				return true, msg, nil
			}
		}
	}
	return false, "本轮未发现可预约空座", nil
}

func (s *Service) tickTomorrow(ctx context.Context, userID uint64, tpl do.ProtocolTemplatesResponse, cookie string, plan do.TomorrowStartRequest) (bool, string, error) {
	if !plan.ExecuteImmediately {
		if wait, msg := scheduledWait(plan.ScheduledStart); wait {
			return false, msg, nil
		}
	}
	if err := s.api(ctx, userID).WarmUpTomorrow(ctx, tpl, cookie, plan.LibraryID); err != nil {
		return false, "", err
	}
	if err := s.api(ctx, userID).SaveTomorrow(ctx, tpl, cookie, plan.LibraryID, plan.Seat.SeatKey); err != nil {
		return false, "", err
	}
	msg := plan.Seat.SeatName + " 明日预约已提交"
	s.notifySuccess(ctx, userID, consts.TaskKindTomorrow, plan.LibraryName, plan.Seat.SeatName, msg)
	return true, msg, nil
}

func (s *Service) failRun(ctx context.Context, run *entity.TaskRun, msg string) error {
	now := time.Now().UTC()
	run.State = stateFailed
	run.Message = msg
	run.LastUpdatedAt = &now
	s.notifyFailure(ctx, run.UserID, run.Kind, msg)
	return dao.UpsertTaskRun(ctx, run)
}

func (s *Service) bumpSuccess(ctx context.Context, userID uint64) {
	cur, _ := dao.GetDashboardMetrics(ctx, userID)
	n := 1
	sec := int64(0)
	if cur != nil {
		n = cur.HistoricalSuccessCount + 1
		sec = cur.TotalGuardSeconds
	}
	_ = dao.UpsertDashboardMetrics(ctx, &entity.DashboardMetrics{
		UserID:                 userID,
		HistoricalSuccessCount: n,
		TotalGuardSeconds:      sec,
	})
}

func toStatus(r entity.TaskRun) do.CoordinatorStatus {
	st := do.CoordinatorStatus{
		Kind:         r.Kind,
		State:        r.State,
		Title:        r.Title,
		Message:      r.Message,
		PollCount:    r.PollCount,
		RequestCount: r.RequestCount,
		Reason:       r.Reason,
		IsActive:     r.State == stateRunning,
	}
	if r.StartedAt != nil {
		st.StartedAt = r.StartedAt.UTC().Format(time.RFC3339)
	}
	if r.LastUpdatedAt != nil {
		st.LastUpdatedAt = r.LastUpdatedAt.UTC().Format(time.RFC3339)
	}
	if r.LastRequestAt != nil {
		st.LastRequestAt = r.LastRequestAt.UTC().Format(time.RFC3339)
	}
	return st
}

func (s *Service) waitTick(ctx context.Context, kind, planJSON string) error {
	return sleepCtx(ctx, tickDelay(kind, planJSON))
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func tickDelay(kind, planJSON string) time.Duration {
	switch kind {
	case consts.TaskKindGrab:
		var plan do.GrabStartRequest
		_ = json.Unmarshal([]byte(planJSON), &plan)
		if wait, _ := scheduledWait(plan.ScheduledStart); wait {
			return minDuration(schedulePreWaitInterval, timeUntilClock(plan.ScheduledStart))
		}
		minMS, maxMS := grabDelayMS(plan)
		if maxMS < minMS {
			maxMS = minMS
		}
		if maxMS == minMS {
			return time.Duration(minMS) * time.Millisecond
		}
		return time.Duration(minMS+rand.IntN(maxMS-minMS+1)) * time.Millisecond //nolint:gosec // jitter delay does not require crypto security
	case consts.TaskKindOccupy:
		var plan do.OccupyStartRequest
		_ = json.Unmarshal([]byte(planJSON), &plan)
		if plan.CheckIntervalMode == "random_ten_to_twenty_seconds" {
			return time.Duration(occupyBaseIntervalSec+rand.IntN(occupyJitterRangeSec)) * time.Second //nolint:gosec // jitter delay does not require crypto security
		}
		return defaultPollInterval
	case consts.TaskKindGlobalLeak:
		var plan do.GlobalLeakStartRequest
		_ = json.Unmarshal([]byte(planJSON), &plan)
		if plan.ScanIntervalSeconds > 0 {
			return time.Duration(plan.ScanIntervalSeconds) * time.Second
		}
		return defaultPollInterval
	case consts.TaskKindTomorrow:
		var plan do.TomorrowStartRequest
		_ = json.Unmarshal([]byte(planJSON), &plan)
		if !plan.ExecuteImmediately {
			if wait, _ := scheduledWait(plan.ScheduledStart); wait {
				return minDuration(schedulePreWaitInterval, timeUntilClock(plan.ScheduledStart))
			}
		}
		return time.Second
	default:
		return time.Second
	}
}

func grabDelayMS(plan do.GrabStartRequest) (minMS, maxMS int) {
	if plan.PollingMinDelayMS > 0 {
		minMS = plan.PollingMinDelayMS
		maxMS = plan.PollingMaxDelayMS
		if maxMS < minMS {
			maxMS = minMS
		}
		return minMS, maxMS
	}
	switch plan.PollingMode {
	case "aggressive":
		return aggressiveDelayMS, aggressiveDelayMS
	case "randomized":
		return randomMinDelayMS, randomMaxDelayMS
	default:
		return defaultDelayMS, defaultDelayMS
	}
}

func scheduledWait(clock string) (bool, string) {
	if clock == "" {
		return false, ""
	}
	until := timeUntilClock(clock)
	if until > 0 {
		return true, "等待定时启动 " + clock
	}
	return false, ""
}

func timeUntilClock(clock string) time.Duration {
	fire, ok := parseClock(clock)
	if !ok {
		return 0
	}
	d := time.Until(fire)
	if d < 0 {
		return 0
	}
	return d
}

func parseClock(clock string) (time.Time, bool) {
	now := time.Now()
	for _, layout := range []string{time.RFC3339, "15:04:05", "15:04"} {
		t, err := time.ParseInLocation(layout, clock, now.Location())
		if err != nil {
			t, err = time.Parse(layout, clock)
		}
		if err != nil {
			continue
		}
		if layout == time.RFC3339 {
			return t, true
		}
		fire := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, now.Location())
		if !fire.After(now) {
			fire = fire.Add(dayDuration)
		}
		return fire, true
	}
	return time.Time{}, false
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func kindTitle(kind string) string {
	switch kind {
	case consts.TaskKindGrab:
		return "抢座"
	case consts.TaskKindOccupy:
		return "占座"
	case consts.TaskKindGlobalLeak:
		return "全域捡漏"
	case consts.TaskKindTomorrow:
		return "明日预约"
	default:
		return kind
	}
}
