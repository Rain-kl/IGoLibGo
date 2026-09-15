// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts_test

import (
	"context"
	"testing"

	"Wavelet/core/contracts"
)

func TestBotCommandRegistry_isInterface(t *testing.T) {
	var _ contracts.BotCommandRegistry = stubRegistry{}
	var _ contracts.BotCommand = stubCmd{}
	var _ contracts.BotConversation = stubConv{}
}

type stubRegistry struct{}

func (stubRegistry) Register(cmd contracts.BotCommand) error { return nil }

func (stubRegistry) RegisterConversation(conv contracts.BotConversation) error {
	return nil
}

type stubCmd struct{}

func (stubCmd) Name() string        { return "" }
func (stubCmd) Aliases() []string   { return nil }
func (stubCmd) Description() string { return "" }
func (stubCmd) Usage() string       { return "" }
func (stubCmd) Handle(ctx context.Context, req contracts.BotCommandRequest) error {
	return nil
}

type stubConv struct{}

func (stubConv) Name() string { return "" }
func (stubConv) OnMessage(ctx context.Context, req contracts.BotConversationRequest) error {
	return nil
}
func (stubConv) OnCancel(ctx context.Context, req contracts.BotConversationRequest) error {
	return nil
}
