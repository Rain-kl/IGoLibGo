// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginAuth_InvalidStateEnds(t *testing.T) {
	conv := &LoginAuthConversation{api: &stubPipeline{}}
	req := &fakeConvReq{userID: 7, stateErr: errors.New("missing")}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.True(t, req.ended)
	assert.Contains(t, req.replies[0], "会话已失效")
}

func TestLoginAuth_TransitionsToCheckin(t *testing.T) {
	api := &stubPipeline{run: &do.PipelineRunResult{NeedAuth: needAuthCheckin, AuthURL: "https://auth.example/checkin"}}
	conv := &LoginAuthConversation{api: api}
	req := &fakeConvReq{userID: 7, text: "code=abc", state: authState{ConfigID: "myseat01"}}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.Equal(t, convCheckinAuth, req.transition)
	assert.False(t, req.ended)
	require.NotNil(t, api.lastReq)
	assert.Equal(t, "code=abc", api.lastReq.Cookie)
	assert.Contains(t, req.replies[0], "登录凭据已更新")
	assert.Contains(t, req.replies[0], "https://auth.example/checkin")
}

func TestLoginAuth_SuccessEnds(t *testing.T) {
	conv := &LoginAuthConversation{api: &stubPipeline{run: &do.PipelineRunResult{Success: true, ConfigID: "myseat01", Name: "主馆"}}}
	req := &fakeConvReq{userID: 7, text: "cookie", state: authState{ConfigID: "myseat01"}}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.True(t, req.ended)
	assert.Contains(t, req.replies[0], "执行成功")
}

func TestLoginAuth_ConfigGone(t *testing.T) {
	conv := &LoginAuthConversation{api: &stubPipeline{runErr: consts.NewError(http.StatusNotFound, consts.CodeNotFound, "sql: no rows")}}
	req := &fakeConvReq{userID: 7, text: "cookie", state: authState{ConfigID: "myseat01"}}
	require.NoError(t, conv.OnMessage(context.Background(), req))
	assert.True(t, req.ended)
	assert.Contains(t, req.replies[0], "已不存在")
	assert.Contains(t, req.replies[0], "myseat01")
	assert.NotContains(t, req.replies[0], "sql:")
}
