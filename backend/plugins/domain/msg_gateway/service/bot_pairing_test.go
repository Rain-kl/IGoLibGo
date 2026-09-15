// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"Wavelet/pkg/idgen"
	"Wavelet/plugins/domain/msg_gateway/consts"
	"Wavelet/plugins/domain/msg_gateway/dao"
	"Wavelet/plugins/domain/msg_gateway/model/do"
	"Wavelet/plugins/domain/msg_gateway/model/entity"
	"Wavelet/plugins/domain/msg_gateway/service"
	"context"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGenerateCode_AlphabetAndLength(t *testing.T) {
	code, err := service.GenerateCode()
	require.NoError(t, err)
	assert.Len(t, code, consts.CodeLength)
	for _, r := range code {
		assert.Contains(t, consts.CodeAlphabet, string(r))
	}
}

func TestNormalizeAndFormat(t *testing.T) {
	assert.Equal(t, "ABCDEFGH", service.NormalizeCode("ab-cd-ef-gh"))
	assert.Equal(t, "ABCD-EFGH", service.FormatCode("ABCDEFGH"))
}

type testDBService struct{ db *gorm.DB }

func (m *testDBService) GORM() *gorm.DB                  { return m.db }
func (m *testDBService) DB(ctx context.Context) *gorm.DB { return m.db.WithContext(ctx) }
func (m *testDBService) Named(_ string) *gorm.DB         { return m.db }

func TestHandleInboundMessage_UnboundUser(t *testing.T) {
	_ = idgen.Init(1)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "pairing_test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}, &entity.MessagePairingCode{}))
	dao.SetDBServiceForTest(&testDBService{db: db})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })

	ctx := context.Background()
	var sentText string
	var sentChannel uint64
	sendFn := func(_ context.Context, chID uint64, _ do.Recipient, text string) error {
		sentChannel = chID
		sentText = text
		return nil
	}

	inbound := do.InboundMessage{
		ChannelID:      10,
		PlatformUserID: "tg_user_999",
		ChatID:         "999",
		Text:           "/start",
	}

	err = service.HandleInboundMessage(ctx, inbound, sendFn)
	require.NoError(t, err)
	assert.Equal(t, uint64(10), sentChannel)
	assert.Contains(t, sentText, "欢迎使用 Wavelet 机器人！")
	assert.Contains(t, sentText, "您的绑定配对码为：")

	// Send another text message (e.g. "hello") -> should reuse existing unexpired pairing code
	var sentText2 string
	sendFn2 := func(_ context.Context, _ uint64, _ do.Recipient, text string) error {
		sentText2 = text
		return nil
	}
	inbound2 := do.InboundMessage{
		ChannelID:      10,
		PlatformUserID: "tg_user_999",
		ChatID:         "999",
		Text:           "hello world",
	}
	err = service.HandleInboundMessage(ctx, inbound2, sendFn2)
	require.NoError(t, err)
	assert.Equal(t, sentText, sentText2) // Reused same pairing code!
}

func TestHandleInboundMessage_BoundUser(t *testing.T) {
	_ = idgen.Init(1)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "pairing_bound_test.db")), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.MessageChannel{}, &entity.MessageBinding{}, &entity.MessagePairingCode{}))
	dao.SetDBServiceForTest(&testDBService{db: db})
	t.Cleanup(func() { dao.SetDBServiceForTest(nil) })

	ctx := context.Background()
	binding := entity.MessageBinding{
		UserID:         1,
		ChannelID:      10,
		PlatformUserID: "tg_bound_user",
	}
	require.NoError(t, dao.CreateMessageBinding(ctx, &binding))

	var sentText string
	sendFn := func(_ context.Context, _ uint64, _ do.Recipient, text string) error {
		sentText = text
		return nil
	}

	inbound := do.InboundMessage{
		ChannelID:      10,
		PlatformUserID: "tg_bound_user",
		ChatID:         "888",
		Text:           "/start",
	}

	err = service.HandleInboundMessage(ctx, inbound, sendFn)
	require.NoError(t, err)
	assert.Contains(t, sentText, "您的账号已成功绑定 Wavelet 平台")
}
