// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"testing"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeReq struct {
	replies     []string
	hasConv     bool
	cancelCalls int
	userID      uint64
	inbound     contracts.BotInbound
}

var _ contracts.BotCommandRequest = (*fakeReq)(nil)

func (r *fakeReq) UserID() uint64                         { return r.userID }
func (r *fakeReq) Args() []string                         { return nil }
func (r *fakeReq) Text() string                           { return "" }
func (r *fakeReq) Inbound() contracts.BotInbound          { return r.inbound }
func (r *fakeReq) Begin(string, any, time.Duration) error { return nil }
func (r *fakeReq) HasConversation() bool                  { return r.hasConv }
func (r *fakeReq) Reply(text string) error {
	r.replies = append(r.replies, text)
	return nil
}
func (r *fakeReq) CancelActive() error {
	r.cancelCalls++
	return nil
}

type stubShow struct{}

func (stubShow) Name() string        { return "show" }
func (stubShow) Aliases() []string   { return nil }
func (stubShow) Description() string { return "列出配置" }
func (stubShow) Usage() string       { return "/show" }
func (stubShow) Handle(context.Context, contracts.BotCommandRequest) error {
	return nil
}

func TestHelpCommand_ListsBuiltinAndBusiness(t *testing.T) {
	cmd := &HelpCommand{List: func() []service.CommandMeta {
		return []service.CommandMeta{
			{Name: "help", Aliases: []string{"start"}, Description: "查看全部可用指令", Usage: "/help", Builtin: true},
			{Name: "cancel", Description: "取消当前进行中的操作", Usage: "/cancel", Builtin: true},
			{Name: "show", Description: "列出配置", Usage: "/show", Builtin: false},
		}
	}}
	req := &fakeReq{}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "/start, /help")
	assert.Contains(t, req.replies[0], "/show")
	assert.Contains(t, req.replies[0], "系统指令")
	assert.Contains(t, req.replies[0], "业务指令")
	assert.NotContains(t, req.replies[0], "/me")
}

func TestHelpCommand_OmitsBusinessSectionWhenNone(t *testing.T) {
	cmd := &HelpCommand{List: func() []service.CommandMeta {
		return []service.CommandMeta{
			{Name: "help", Aliases: []string{"start"}, Description: "查看全部可用指令", Usage: "/help", Builtin: true},
			{Name: "cancel", Description: "取消当前进行中的操作", Usage: "/cancel", Builtin: true},
		}
	}}
	req := &fakeReq{}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "系统指令")
	assert.Contains(t, req.replies[0], "/start, /help")
	assert.Contains(t, req.replies[0], "/cancel")
	assert.NotContains(t, req.replies[0], "业务指令")
}

func TestHelpCommand_StartIsHelpAlias(t *testing.T) {
	cmd := &HelpCommand{}
	assert.Equal(t, consts.CommandHelp, cmd.Name())
	assert.Equal(t, []string{consts.CommandStart}, cmd.Aliases())
	assert.Equal(t, "查看全部可用指令", cmd.Description())
	assert.Equal(t, "/help", cmd.Usage())
}

func TestHelpCommand_FromRegistryIncludesMe(t *testing.T) {
	reg := service.NewBotRegistry()
	require.NoError(t, RegisterBuiltins(reg))
	require.NoError(t, reg.Register(stubShow{}))

	help, builtin, allowUnbound, ok := reg.LookupCommand(consts.CommandHelp)
	require.True(t, ok)
	assert.True(t, builtin)
	assert.True(t, allowUnbound)

	start, _, _, ok := reg.LookupCommand(consts.CommandStart)
	require.True(t, ok)
	assert.Equal(t, help.Name(), start.Name())

	_, meBuiltin, meUnbound, ok := reg.LookupCommand(consts.CommandMe)
	require.True(t, ok)
	assert.True(t, meBuiltin)
	assert.True(t, meUnbound)

	req := &fakeReq{}
	require.NoError(t, help.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "/start, /help")
	assert.Contains(t, req.replies[0], "/show")
	assert.Contains(t, req.replies[0], "/cancel")
	assert.Contains(t, req.replies[0], "/me")
}

func TestCancelCommand_NoConversation(t *testing.T) {
	cmd := &CancelCommand{}
	req := &fakeReq{}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Equal(t, "当前没有进行中的操作", req.replies[0])
	assert.NotContains(t, req.replies[0], "配对码")
	assert.Zero(t, req.cancelCalls)
}

func TestCancelCommand_CancelsActive(t *testing.T) {
	cmd := &CancelCommand{}
	req := &fakeReq{hasConv: true}
	require.NoError(t, cmd.Handle(context.Background(), req))
	assert.Equal(t, 1, req.cancelCalls)
	require.Len(t, req.replies, 1)
	assert.Equal(t, "已取消当前操作", req.replies[0])
}

func TestRegisterBuiltins(t *testing.T) {
	reg := service.NewBotRegistry()
	require.NoError(t, RegisterBuiltins(reg))

	list := reg.List()
	require.Len(t, list, 3)
	assert.Equal(t, consts.CommandHelp, list[0].Name)
	assert.Equal(t, []string{consts.CommandStart}, list[0].Aliases)
	assert.True(t, list[0].Builtin)
	assert.Equal(t, consts.CommandCancel, list[1].Name)
	assert.True(t, list[1].Builtin)
	assert.Equal(t, consts.CommandMe, list[2].Name)
	assert.True(t, list[2].Builtin)

	_, builtin, allowUnbound, ok := reg.LookupCommand(consts.CommandCancel)
	require.True(t, ok)
	assert.True(t, builtin)
	assert.True(t, allowUnbound)

	_, builtin, allowUnbound, ok = reg.LookupCommand(consts.CommandMe)
	require.True(t, ok)
	assert.True(t, builtin)
	assert.True(t, allowUnbound)

	cancel := &CancelCommand{}
	assert.Equal(t, consts.CommandCancel, cancel.Name())
	assert.Empty(t, cancel.Aliases())
	assert.Equal(t, "取消当前进行中的操作", cancel.Description())
	assert.Equal(t, "/cancel", cancel.Usage())
}
