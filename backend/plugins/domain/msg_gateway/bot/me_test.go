// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"testing"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/model/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeCommand_Meta(t *testing.T) {
	cmd := &MeCommand{}
	assert.Equal(t, consts.CommandMe, cmd.Name())
	assert.Empty(t, cmd.Aliases())
	assert.Equal(t, "查看绑定状态与个人信息", cmd.Description())
	assert.Equal(t, "/me", cmd.Usage())
}

func TestMeCommand_UnboundUsesUnboundReply(t *testing.T) {
	called := false
	cmd := &MeCommand{
		UnboundReply: func(_ context.Context, req contracts.BotCommandRequest) error {
			called = true
			return req.Reply("您尚未绑定，请使用配对码完成绑定")
		},
	}
	req := &fakeReq{}
	require.NoError(t, cmd.Handle(context.Background(), req))
	assert.True(t, called)
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "尚未绑定")
}

func TestMeCommand_BoundShowsIdentityOmitsSecrets(t *testing.T) {
	cmd := &MeCommand{
		ListBindings: func(_ context.Context, userID uint64) ([]entity.MessageBinding, error) {
			return []entity.MessageBinding{
				{ChannelID: 10, PlatformUserID: "tg_123", UserID: userID},
				{ChannelID: 20, PlatformUserID: "qq_456", UserID: userID},
			}, nil
		},
		GetChannel: func(_ context.Context, id uint64) (*entity.MessageChannel, error) {
			switch id {
			case 10:
				return &entity.MessageChannel{ID: 10, Name: "Telegram 主频道", Type: consts.ChannelTypeTelegram}, nil
			case 20:
				return &entity.MessageChannel{ID: 20, Name: "QQ 工作频道", Type: consts.ChannelTypeQQ}, nil
			default:
				return nil, consts.ErrRecordNotFound
			}
		},
		GetUser: func(_ context.Context, id uint64) (*contracts.UserDTO, error) {
			return &contracts.UserDTO{
				ID:       id,
				Username: "alice",
				Nickname: "Alice",
				Email:    "alice@example.com",
				Phone:    "13800138000",
			}, nil
		},
	}
	req := &fakeReq{
		userID: 42,
		inbound: contracts.BotInbound{
			ChannelID:      10,
			ChannelType:    consts.ChannelTypeTelegram,
			PlatformUserID: "tg_123",
		},
	}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	text := req.replies[0]
	assert.Contains(t, text, "alice")
	assert.Contains(t, text, "Alice")
	assert.Contains(t, text, "42")
	assert.Contains(t, text, consts.ChannelTypeTelegram)
	assert.Contains(t, text, "Telegram 主频道")
	assert.Contains(t, text, "tg_123")
	assert.Contains(t, text, "QQ 工作频道")
	assert.NotRegexp(t, `\S+@\S+`, text)
	assert.NotContains(t, text, "cookie")
	assert.NotContains(t, text, "Cookie")
	assert.NotContains(t, text, "alice@example.com")
	assert.NotContains(t, text, "13800138000")
}

func TestMeCommand_BoundWithoutUserServiceShowsID(t *testing.T) {
	cmd := &MeCommand{
		ListBindings: func(_ context.Context, userID uint64) ([]entity.MessageBinding, error) {
			return []entity.MessageBinding{
				{ChannelID: 10, PlatformUserID: "tg_123", UserID: userID},
			}, nil
		},
		GetChannel: func(_ context.Context, id uint64) (*entity.MessageChannel, error) {
			return &entity.MessageChannel{ID: id, Name: "Telegram 主频道", Type: consts.ChannelTypeTelegram}, nil
		},
	}
	req := &fakeReq{
		userID: 7,
		inbound: contracts.BotInbound{
			ChannelID:      10,
			ChannelType:    consts.ChannelTypeTelegram,
			PlatformUserID: "tg_123",
		},
	}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	text := req.replies[0]
	assert.Contains(t, text, "7")
	assert.Contains(t, text, consts.ChannelTypeTelegram)
	assert.NotContains(t, text, "@")
	assert.NotContains(t, text, "cookie")
}
