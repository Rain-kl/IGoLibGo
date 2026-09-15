// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"net/http"
	"testing"

	"Wavelet/igo-lib/plugins/igo/consts"
	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand_RequiresConfigID(t *testing.T) {
	cmd := &RunCommand{api: &stubPipeline{}}
	req := &fakeReq{userID: 7}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.NotEmpty(t, req.replies)
	assert.Contains(t, req.replies[0], "/run")
	assert.Empty(t, req.begun)
}

func TestRunCommand_NotFound(t *testing.T) {
	cmd := &RunCommand{api: &stubPipeline{runErr: consts.NewError(http.StatusNotFound, consts.CodeNotFound, "missing")}}
	req := &fakeReq{userID: 7, args: []string{"nope"}}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.NotEmpty(t, req.replies)
	assert.Contains(t, req.replies[0], "nope")
	assert.NotContains(t, req.replies[0], "missing")
}

func TestRunCommand_BeginsLoginAuth(t *testing.T) {
	cmd := &RunCommand{api: &stubPipeline{run: &do.PipelineRunResult{
		NeedAuth: needAuthLogin,
		AuthURL:  "https://auth.example/login",
		ConfigID: "myseat01",
	}}}
	req := &fakeReq{userID: 7, args: []string{"myseat01"}}
	require.NoError(t, cmd.Handle(context.Background(), req))
	assert.Equal(t, convLoginAuth, req.begun)
	require.NotEmpty(t, req.replies)
	assert.Contains(t, req.replies[0], "https://auth.example/login")
	assert.Contains(t, req.replies[0], "/cancel")
}

func TestRunCommand_BeginsCheckinAuth(t *testing.T) {
	cmd := &RunCommand{api: &stubPipeline{run: &do.PipelineRunResult{NeedAuth: needAuthCheckin, AuthURL: "https://auth.example/checkin"}}}
	req := &fakeReq{userID: 7, args: []string{"myseat01"}}
	require.NoError(t, cmd.Handle(context.Background(), req))
	assert.Equal(t, convCheckinAuth, req.begun)
	assert.Contains(t, req.replies[0], "https://auth.example/checkin")
}

func TestRunCommand_Success(t *testing.T) {
	cmd := &RunCommand{api: &stubPipeline{run: &do.PipelineRunResult{
		Success: true, Name: "主馆", ConfigID: "myseat01", ReservationStatus: "ok", ExecutedAt: "now",
	}}}
	req := &fakeReq{userID: 7, args: []string{"myseat01"}}
	require.NoError(t, cmd.Handle(context.Background(), req))
	assert.Empty(t, req.begun)
	assert.Contains(t, req.replies[0], "执行成功")
	assert.Contains(t, req.replies[0], "myseat01")
}
