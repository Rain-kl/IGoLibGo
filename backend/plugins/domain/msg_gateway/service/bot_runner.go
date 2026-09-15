// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"Wavelet/pkg/logger"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
	"context"
	"fmt"
	"sync"
)

// Handler processes one inbound message.
type Handler func(ctx context.Context, msg do.InboundMessage) error

// Factory constructs a Channel from decrypted config.
type Factory func(cfg do.ChannelConfig, onInbound Handler) (Channel, error)

// Channel is one connected messaging adapter.
type Channel interface {
	Type() string
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Send(ctx context.Context, to do.Recipient, msg do.OutboundMessage) error
	Capabilities() do.Capability
}

var (
	factoriesMu sync.RWMutex
	factories   = map[string]Factory{}
)

// Register stores a channel factory under typ.
func Register(typ string, fn Factory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	factories[typ] = fn
}

// Lookup returns a previously registered factory.
func Lookup(typ string) (Factory, bool) {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()
	fn, ok := factories[typ]
	return fn, ok
}

// Note: Bot Runner 生命周期管理 — 见 .agents/notes/implemented/bug-fix/2026-09-15-telegram-bot-start-reply.md

// Runner manages lifecycle for long-lived channel adapters (WebSocket, long-polling, etc.).
type Runner struct {
	mu       sync.RWMutex
	running  bool
	cancel   context.CancelFunc
	channels map[uint64]Channel
}

// GlobalRunner is the default global runner instance.
var GlobalRunner = &Runner{}

// Start starts all background long-lived channel runners.
func Start(ctx context.Context) error {
	return GlobalRunner.Start(ctx)
}

// Stop stops the channel runner.
func Stop() {
	GlobalRunner.Stop()
}

// Reload restarts active channels to pick up DB changes.
func Reload(ctx context.Context) error {
	return GlobalRunner.Reload(ctx)
}

// Start loads enabled channels and starts long-polling or WebSocket connections.
func (r *Runner) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}

	runCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	r.running = true
	r.channels = make(map[uint64]Channel)
	r.mu.Unlock()

	logger.InfoF(runCtx, "[MessageGateway] Starting bot channel runners...")
	return r.syncChannels(runCtx)
}

func (r *Runner) syncChannels(ctx context.Context) error {
	rows, err := dao.ListEnabledMessageChannels(ctx)
	if err != nil {
		logger.ErrorF(ctx, "[MessageGateway] List enabled channels error: %v", err)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range rows {
		row := &rows[i]
		if _, exists := r.channels[row.ID]; exists {
			continue
		}
		factory, ok := Lookup(row.Type)
		if !ok {
			logger.WarnF(ctx, "[MessageGateway] Channel type %s not registered for channel %d", row.Type, row.ID)
			continue
		}
		cfg, err := channelConfigFromRow(row)
		if err != nil {
			logger.ErrorF(ctx, "[MessageGateway] Channel config decode error channel=%d: %v", row.ID, err)
			continue
		}
		ch, err := factory(cfg, func(inboundCtx context.Context, msg do.InboundMessage) error {
			if reg := GetBotRegistry(); reg != nil {
				return reg.Dispatch(inboundCtx, msg, r.SendText)
			}
			return HandleInboundMessage(inboundCtx, msg, r.SendText)
		})
		if err != nil {
			logger.ErrorF(ctx, "[MessageGateway] Factory construct error channel=%d: %v", row.ID, err)
			continue
		}
		if err := ch.Connect(ctx); err != nil {
			logger.ErrorF(ctx, "[MessageGateway] Connect error channel=%d: %v", row.ID, err)
			continue
		}
		r.channels[row.ID] = ch
		logger.InfoF(ctx, "[MessageGateway] Connected bot channel %d (%s: %s)", row.ID, row.Type, row.Name)
	}
	return nil
}

// SendText sends text through an active connected channel.
func (r *Runner) SendText(ctx context.Context, channelID uint64, to do.Recipient, text string) error {
	r.mu.RLock()
	ch, ok := r.channels[channelID]
	r.mu.RUnlock()

	if !ok || ch == nil {
		return fmt.Errorf("channel %d not connected", channelID)
	}
	return ch.Send(ctx, to, do.OutboundMessage{Text: text})
}

// Stop disconnects all running bot channels.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}
	if r.cancel != nil {
		r.cancel()
	}
	for id, ch := range r.channels {
		_ = ch.Disconnect(context.Background())
		delete(r.channels, id)
	}
	r.running = false
}

// Reload disconnects existing channels and re-synchronizes with the database.
func (r *Runner) Reload(ctx context.Context) error {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return nil
	}
	for id, ch := range r.channels {
		_ = ch.Disconnect(ctx)
		delete(r.channels, id)
	}
	r.mu.Unlock()

	return r.syncChannels(ctx)
}
