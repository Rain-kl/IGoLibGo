// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/pkg/logger"
	"Wavelet/plugins/domain/msg_gateway/model/do"
)

// Note: 对话占位按 (channelID, platformUserID) 排他、过期只靠缓存 TTL — 见 .agents/notes/implemented/architecture/2026-09-15-bot-command-registry.md

type conversationRecord struct {
	Name     string          `json:"name"`
	State    json.RawMessage `json:"state"`
	TTLNanos int64           `json:"ttl_nanos"`
	ExpireAt int64           `json:"expire_at"`
}

func occupancyKey(channelID uint64, platformUserID string) string {
	return fmt.Sprintf("msg_gateway:bot_conv:%d:%s", channelID, platformUserID)
}

func loadOccupancy(ctx context.Context, channelID uint64, platformUserID string) (*conversationRecord, bool) {
	cache := GetCache(ctx)
	if cache == nil {
		return nil, false
	}
	var rec conversationRecord
	if err := cache.Get(ctx, occupancyKey(channelID, platformUserID), &rec); err != nil {
		if !errors.Is(err, contracts.ErrCacheMiss) {
			logger.ErrorF(ctx, "bot inbound: load occupancy channel=%d platform_user=%s: %v", channelID, platformUserID, err)
		}
		return nil, false
	}
	if rec.ExpireAt > 0 && time.Now().UnixNano() >= rec.ExpireAt {
		if err := cache.Delete(ctx, occupancyKey(channelID, platformUserID)); err != nil {
			logger.ErrorF(ctx, "bot inbound: drop expired occupancy: %v", err)
		}
		return nil, false
	}
	return &rec, true
}

func saveOccupancy(ctx context.Context, channelID uint64, platformUserID string, rec *conversationRecord, ttl time.Duration) error {
	cache := GetCache(ctx)
	if cache == nil {
		return fmt.Errorf("bot conversation cache unavailable")
	}
	return cache.Set(ctx, occupancyKey(channelID, platformUserID), rec, ttl)
}

func deleteOccupancy(ctx context.Context, channelID uint64, platformUserID string) error {
	cache := GetCache(ctx)
	if cache == nil {
		return nil
	}
	return cache.Delete(ctx, occupancyKey(channelID, platformUserID))
}

func occupancyTTL(rec *conversationRecord) time.Duration {
	if rec == nil {
		return 0
	}
	if rec.ExpireAt > 0 {
		remain := time.Until(time.Unix(0, rec.ExpireAt))
		if remain > 0 {
			return remain
		}
	}
	return time.Duration(rec.TTLNanos)
}

func (r *commandRequest) Begin(name string, state any, ttl time.Duration) error {
	if GetCache(r.ctx) == nil {
		return fmt.Errorf("bot conversation cache unavailable")
	}
	if r.reg == nil {
		return fmt.Errorf("bot conversation %q is not registered", name)
	}
	conv, ok := r.reg.LookupConversation(name)
	if !ok {
		return fmt.Errorf("bot conversation %q is not registered", name)
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("bot conversation state: %w", err)
	}
	rec := &conversationRecord{
		Name:     conv.Name(),
		State:    raw,
		TTLNanos: ttl.Nanoseconds(),
	}
	if ttl > 0 {
		rec.ExpireAt = time.Now().Add(ttl).UnixNano()
	}
	return saveOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID, rec, ttl)
}

func (r *commandRequest) HasConversation() bool {
	_, ok := loadOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID)
	return ok
}

func (r *commandRequest) CancelActive() error {
	rec, ok := loadOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID)
	if !ok {
		return nil
	}
	if r.reg != nil {
		if conv, found := r.reg.LookupConversation(rec.Name); found {
			req := newConversationRequest(r.ctx, r.reg, r.userID, r.text, r.in, r.send, rec)
			r.reg.invokeOnCancel(r.ctx, conv, req)
		}
	}
	return deleteOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID)
}

type conversationRequest struct {
	ctx    context.Context
	reg    *BotRegistry
	userID uint64
	text   string
	in     contracts.BotInbound
	send   SendTextFn
	rec    *conversationRecord
}

var _ contracts.BotConversationRequest = (*conversationRequest)(nil)

func newConversationRequest(ctx context.Context, reg *BotRegistry, userID uint64, text string, in contracts.BotInbound, send SendTextFn, rec *conversationRecord) *conversationRequest {
	return &conversationRequest{ctx: ctx, reg: reg, userID: userID, text: text, in: in, send: send, rec: rec}
}

func (r *conversationRequest) UserID() uint64                { return r.userID }
func (r *conversationRequest) Text() string                  { return r.text }
func (r *conversationRequest) Inbound() contracts.BotInbound { return r.in }

func (r *conversationRequest) Reply(text string) error {
	if r.send == nil {
		return nil
	}
	return r.send(r.ctx, r.in.ChannelID, doRecipient(r.in), text)
}

func (r *conversationRequest) State(dst any) error {
	if r.rec == nil || len(r.rec.State) == 0 {
		return nil
	}
	return json.Unmarshal(r.rec.State, dst)
}

func (r *conversationRequest) SetState(v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("bot conversation state: %w", err)
	}
	if r.rec == nil {
		r.rec = &conversationRecord{}
	}
	r.rec.State = raw
	return saveOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID, r.rec, occupancyTTL(r.rec))
}

func (r *conversationRequest) Transition(conversation string, state any) error {
	if r.reg == nil {
		return fmt.Errorf("bot conversation cache unavailable")
	}
	conv, ok := r.reg.LookupConversation(conversation)
	if !ok {
		return fmt.Errorf("bot conversation %q is not registered", conversation)
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("bot conversation state: %w", err)
	}
	if r.rec == nil {
		r.rec = &conversationRecord{}
	}
	ttl := time.Duration(r.rec.TTLNanos)
	r.rec.Name = conv.Name()
	r.rec.State = raw
	if ttl > 0 {
		r.rec.ExpireAt = time.Now().Add(ttl).UnixNano()
	}
	return saveOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID, r.rec, ttl)
}

func (r *conversationRequest) End() error {
	return deleteOccupancy(r.ctx, r.in.ChannelID, r.in.PlatformUserID)
}

func (r *BotRegistry) deliverOccupancy(ctx context.Context, in contracts.BotInbound, send SendTextFn) bool {
	rec, ok := loadOccupancy(ctx, in.ChannelID, in.PlatformUserID)
	if !ok {
		return false
	}
	conv, found := r.LookupConversation(rec.Name)
	if !found {
		if err := deleteOccupancy(ctx, in.ChannelID, in.PlatformUserID); err != nil {
			logger.ErrorF(ctx, "bot inbound: drop stale occupancy %q: %v", rec.Name, err)
		}
		return false
	}
	req := newConversationRequest(ctx, r, in.UserID, in.Text, in, send, rec)
	r.invokeOnMessage(ctx, conv, req)
	return true
}

func (r *BotRegistry) invokeOnMessage(ctx context.Context, conv contracts.BotConversation, req contracts.BotConversationRequest) {
	r.invokeConversation(ctx, conv.Name(), "OnMessage", func() error {
		return conv.OnMessage(ctx, req)
	})
}

func (r *BotRegistry) invokeOnCancel(ctx context.Context, conv contracts.BotConversation, req contracts.BotConversationRequest) {
	r.invokeConversation(ctx, conv.Name(), "OnCancel", func() error {
		return conv.OnCancel(ctx, req)
	})
}

func (r *BotRegistry) invokeConversation(ctx context.Context, name, method string, fn func() error) {
	defer func() {
		if rec := recover(); rec != nil {
			logger.ErrorF(ctx, "bot inbound: conversation %q %s panic: %v\n%s", name, method, rec, debug.Stack())
		}
	}()
	if err := fn(); err != nil {
		logger.ErrorF(ctx, "bot inbound: conversation %q %s: %v", name, method, err)
	}
}

func doRecipient(in contracts.BotInbound) do.Recipient {
	return do.Recipient{ChatID: in.ChatID, PlatformUserID: in.PlatformUserID}
}
