// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package bot

import (
	"context"
	"errors"
	"testing"

	"Wavelet/igo-lib/plugins/igo/model/do"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand_Empty(t *testing.T) {
	cmd := &ShowCommand{api: &stubPipeline{}}
	req := &fakeReq{userID: 7}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "暂无配置")
}

func TestShowCommand_ListsCardsWithoutSecrets(t *testing.T) {
	cmd := &ShowCommand{api: &stubPipeline{list: []do.PipelineConfigDTO{
		{ID: "myseat01", Name: "主馆", LibraryName: "主图书馆", Floor: "2", SeatName: "A1", SeatKey: "seat-1", AutoCheckin: true, HasCookie: true, CookieMasked: "SECRET"},
	}}}
	req := &fakeReq{userID: 7}
	require.NoError(t, cmd.Handle(context.Background(), req))
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "myseat01")
	assert.Contains(t, req.replies[0], "有效")
	assert.Contains(t, req.replies[0], "/run")
	assert.NotContains(t, req.replies[0], "SECRET")
}

func TestShowCommand_ListError(t *testing.T) {
	cmd := &ShowCommand{api: &stubPipeline{listErr: errors.New("db down")}}
	req := &fakeReq{userID: 7}
	err := cmd.Handle(context.Background(), req)
	require.Error(t, err)
	require.Len(t, req.replies, 1)
	assert.Contains(t, req.replies[0], "获取配置列表失败")
	assert.NotContains(t, req.replies[0], "db down")
}
