// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"testing"

	"Wavelet/core/contracts"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ contracts.BotCommandRegistry = (*service.BotRegistry)(nil)

type stubCmd struct {
	name, desc, usage string
	aliases           []string
}

func (s stubCmd) Name() string        { return s.name }
func (s stubCmd) Aliases() []string   { return s.aliases }
func (s stubCmd) Description() string { return s.desc }
func (s stubCmd) Usage() string       { return s.usage }
func (s stubCmd) Handle(context.Context, contracts.BotCommandRequest) error {
	return nil
}

type stubConv struct {
	name string
}

func (s stubConv) Name() string { return s.name }
func (s stubConv) OnMessage(context.Context, contracts.BotConversationRequest) error {
	return nil
}
func (s stubConv) OnCancel(context.Context, contracts.BotConversationRequest) error {
	return nil
}

func validCmd(name string, aliases ...string) stubCmd {
	return stubCmd{name: name, desc: "description for " + name, usage: "/" + name, aliases: aliases}
}

func TestBotRegistry(t *testing.T) {
	t.Run("lookup command is case insensitive", func(t *testing.T) {
		reg := service.NewBotRegistry()
		require.NoError(t, reg.Register(validCmd("show")))

		cmd, builtin, allowUnbound, ok := reg.LookupCommand("Show")
		require.True(t, ok)
		require.NotNil(t, cmd)
		assert.Equal(t, "show", cmd.Name())
		assert.False(t, builtin)
		assert.False(t, allowUnbound)
	})

	t.Run("register rejects invalid commands", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(*testing.T, *service.BotRegistry)
			cmd     contracts.BotCommand
			wantSub string
		}{
			{
				name:    "empty name",
				cmd:     stubCmd{name: "", desc: "d", usage: "/x"},
				wantSub: "name",
			},
			{
				name:    "whitespace name",
				cmd:     stubCmd{name: "  ", desc: "d", usage: "/x"},
				wantSub: "name",
			},
			{
				name:    "empty description",
				cmd:     stubCmd{name: "foo", desc: "", usage: "/foo"},
				wantSub: "foo",
			},
			{
				name:    "empty usage",
				cmd:     stubCmd{name: "foo", desc: "d", usage: ""},
				wantSub: "foo",
			},
			{
				name:    "name contains slash",
				cmd:     stubCmd{name: "foo/bar", desc: "d", usage: "/x"},
				wantSub: "foo/bar",
			},
			{
				name:    "name contains whitespace",
				cmd:     stubCmd{name: "foo bar", desc: "d", usage: "/x"},
				wantSub: "foo bar",
			},
			{
				name: "duplicate name",
				setup: func(t *testing.T, reg *service.BotRegistry) {
					t.Helper()
					require.NoError(t, reg.Register(validCmd("show")))
				},
				cmd:     validCmd("Show"),
				wantSub: "show",
			},
			{
				name: "alias conflicts with builtin name",
				setup: func(t *testing.T, reg *service.BotRegistry) {
					t.Helper()
					require.NoError(t, reg.RegisterBuiltin(validCmd(consts.CommandHelp, consts.CommandStart), true))
				},
				cmd:     validCmd("foo", "HELP"),
				wantSub: "help",
			},
			{
				name:    "aliases collide with each other",
				cmd:     validCmd("foo", "bar", "BAR"),
				wantSub: "bar",
			},
			{
				name:    "alias collides with own name",
				cmd:     validCmd("foo", "FOO"),
				wantSub: "foo",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				reg := service.NewBotRegistry()
				if tt.setup != nil {
					tt.setup(t, reg)
				}
				err := reg.Register(tt.cmd)
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantSub)
			})
		}
	})

	t.Run("conversation name requires dot and is unique", func(t *testing.T) {
		reg := service.NewBotRegistry()

		err := reg.RegisterConversation(stubConv{name: "login_auth"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "login_auth")

		require.NoError(t, reg.RegisterConversation(stubConv{name: "igo.login_auth"}))

		conv, ok := reg.LookupConversation("IGo.Login_Auth")
		require.True(t, ok)
		require.NotNil(t, conv)
		assert.Equal(t, "igo.login_auth", conv.Name())

		err = reg.RegisterConversation(stubConv{name: "IGO.login_auth"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "igo.login_auth")
	})

	t.Run("RegisterBuiltin listed as Builtin", func(t *testing.T) {
		reg := service.NewBotRegistry()
		require.NoError(t, reg.RegisterBuiltin(validCmd(consts.CommandHelp, consts.CommandStart), true))
		require.NoError(t, reg.Register(validCmd("show")))

		list := reg.List()
		require.Len(t, list, 2)

		assert.Equal(t, consts.CommandHelp, list[0].Name)
		assert.Equal(t, []string{consts.CommandStart}, list[0].Aliases)
		assert.True(t, list[0].Builtin)

		assert.Equal(t, "show", list[1].Name)
		assert.False(t, list[1].Builtin)

		cmd, builtin, allowUnbound, ok := reg.LookupCommand("Start")
		require.True(t, ok)
		assert.Equal(t, consts.CommandHelp, cmd.Name())
		assert.True(t, builtin)
		assert.True(t, allowUnbound)
	})
}
