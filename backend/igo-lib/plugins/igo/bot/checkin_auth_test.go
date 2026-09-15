// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"testing"

	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckinAuth_SuccessEnds(t *testing.T) {
	api := &stubPipeline{run: &do.PipelineRunResult{Success: true, ConfigID: "myseat01", Name: "主馆", CheckinStatus: "ok"}}
	conv := &CheckinAuthConversation{api: api}
	req := &fakeConvReq{userID: 7, text: "token", state: authState{ConfigID: "myseat01"}}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.True(t, req.ended)
	require.NotNil(t, api.lastReq)
	assert.Equal(t, "token", api.lastReq.CheckinToken)
	assert.Contains(t, req.replies[0], "签到凭据已更新")
}

func TestCheckinAuth_StillNeededStaysOpen(t *testing.T) {
	conv := &CheckinAuthConversation{api: &stubPipeline{run: &do.PipelineRunResult{NeedAuth: needAuthCheckin}}}
	req := &fakeConvReq{userID: 7, text: "bad", state: authState{ConfigID: "myseat01"}}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.False(t, req.ended)
	assert.Contains(t, req.replies[0], "仍无效")
}

func TestCheckinAuth_OnCancelNoReply(t *testing.T) {
	conv := &CheckinAuthConversation{}
	req := &fakeConvReq{}
	require.NoError(t, conv.OnCancel(context.Background(), req))
	assert.Empty(t, req.replies)
}
