// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/service"
)

var _ pipelineAPI = (*service.Service)(nil)

// Register attaches IGo bot commands and conversations. Panics on conflict so
// process startup fails (ctx.Bind cannot return error).
func Register(reg contracts.BotCommandRegistry, svc *service.Service) {
	if reg == nil || svc == nil {
		return
	}
	mustRegister(reg.Register(&ShowCommand{api: svc}))
	mustRegister(reg.Register(&RunCommand{api: svc}))
	mustRegister(reg.RegisterConversation(&LoginAuthConversation{api: svc}))
	mustRegister(reg.RegisterConversation(&CheckinAuthConversation{api: svc}))
}

func mustRegister(err error) {
	if err != nil {
		panic(err)
	}
}
