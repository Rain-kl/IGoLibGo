// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo/model/do"
)

type fakeReq struct {
	userID  uint64
	args    []string
	text    string
	replies []string
	begun   string
	state   any
}

var _ contracts.BotCommandRequest = (*fakeReq)(nil)

func (r *fakeReq) UserID() uint64                { return r.userID }
func (r *fakeReq) Args() []string                { return r.args }
func (r *fakeReq) Text() string                  { return r.text }
func (r *fakeReq) Inbound() contracts.BotInbound { return contracts.BotInbound{} }
func (r *fakeReq) Reply(text string) error {
	r.replies = append(r.replies, text)
	return nil
}
func (r *fakeReq) Begin(conversation string, state any, _ time.Duration) error {
	r.begun = conversation
	r.state = state
	return nil
}
func (r *fakeReq) HasConversation() bool { return r.begun != "" }
func (r *fakeReq) CancelActive() error   { return nil }

type fakeConvReq struct {
	userID     uint64
	text       string
	state      any
	replies    []string
	ended      bool
	transition string
	stateErr   error
}

var _ contracts.BotConversationRequest = (*fakeConvReq)(nil)

func (r *fakeConvReq) UserID() uint64                { return r.userID }
func (r *fakeConvReq) Text() string                  { return r.text }
func (r *fakeConvReq) Inbound() contracts.BotInbound { return contracts.BotInbound{} }
func (r *fakeConvReq) Reply(text string) error {
	r.replies = append(r.replies, text)
	return nil
}
func (r *fakeConvReq) State(dst any) error {
	if r.stateErr != nil {
		return r.stateErr
	}
	if st, ok := r.state.(authState); ok {
		if ptr, ok := dst.(*authState); ok {
			*ptr = st
			return nil
		}
	}
	return nil
}
func (r *fakeConvReq) SetState(v any) error { r.state = v; return nil }
func (r *fakeConvReq) Transition(conversation string, state any) error {
	r.transition = conversation
	r.state = state
	return nil
}
func (r *fakeConvReq) End() error { r.ended = true; return nil }

type stubPipeline struct {
	list    []do.PipelineConfigDTO
	listErr error
	run     *do.PipelineRunResult
	runErr  error
	lastID  string
	lastReq *do.RunPipelineRequest
}

func (s *stubPipeline) ListPipelineConfigs(context.Context, uint64) ([]do.PipelineConfigDTO, error) {
	return s.list, s.listErr
}

func (s *stubPipeline) RunPipeline(_ context.Context, _ uint64, id string, overrideReq *do.RunPipelineRequest) (*do.PipelineRunResult, error) {
	s.lastID = id
	s.lastReq = overrideReq
	return s.run, s.runErr
}
