// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Wavelet/core/contracts"
	"Wavelet/igo-lib/plugins/igo"
	"Wavelet/igo-lib/plugins/igo/dao"
	"Wavelet/igo-lib/plugins/igo/model/do"
	"Wavelet/igo-lib/plugins/igo/service"
	"Wavelet/pkg/idgen"

	"github.com/glebarez/sqlite"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"io/fs"
	"path/filepath"
)

type testDB struct{ db *gorm.DB }

func (s testDB) GORM() *gorm.DB                  { return s.db }
func (s testDB) DB(ctx context.Context) *gorm.DB { return s.db.WithContext(ctx) }
func (s testDB) Named(string) *gorm.DB           { return s.db }

func setupService(t *testing.T, handler http.HandlerFunc) *service.Service {
	t.Helper()
	require.NoError(t, idgen.Init(1))
	gdb, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "igo.db")), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := gdb.DB()
	require.NoError(t, err)
	sub, err := fs.Sub(igo.MigrationsFS, "migrations/sqlite")
	require.NoError(t, err)
	provider, err := goose.NewProvider(goose.DialectSQLite3, sqlDB, sub)
	require.NoError(t, err)
	_, err = provider.Up(context.Background())
	require.NoError(t, err)
	dao.SetDBService(testDB{db: gdb})
	t.Cleanup(func() { dao.SetDBService((contracts.DBService)(nil)) })

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	svc := service.New()
	svc.SetHTTPClient(&http.Client{Timeout: 5 * time.Second, Transport: rewriteHost(srv.URL)})
	return svc
}

type rewriteHost string

func (h rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	u := string(h)
	base := strings.TrimPrefix(u, "http://")
	req.URL.Scheme = "http"
	req.URL.Host = base
	req.Host = base
	return http.DefaultTransport.RoundTrip(req)
}

func TestAuthenticateFromCookieAndListLibraries(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "graphql") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[
				{"lib_id":8,"lib_name":"二楼","lib_floor":"2","is_open":true,"lib_rt":{"seats_total":20,"seats_used":1,"seats_booking":0}}
			]}}}}`)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "SERVERID", Value: "s1"})
		http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: "tok"})
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	res, err := svc.AuthenticateFromCookie(ctx, 42, do.AuthFromCookieRequest{Cookie: "Authorization=tok; SERVERID=s1", Remember: true})
	require.NoError(t, err)
	require.True(t, res.Session.Authorized)
	require.Len(t, res.Libraries, 1)
	assert.Equal(t, 8, res.Libraries[0].LibraryID)

	sess, err := svc.GetSession(ctx, 42)
	require.NoError(t, err)
	assert.True(t, sess.Authorized)

	libs, err := svc.ListLibraries(ctx, 42)
	require.NoError(t, err)
	require.Len(t, libs, 1)

	raw, _ := json.Marshal(libs)
	assert.Contains(t, string(raw), "二楼")
}

func TestAuthenticateFromCookie_AutoExtractCode(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "graphql") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"data":{"userAuth":{"reserve":{"libs":[
				{"lib_id":8,"lib_name":"二楼","lib_floor":"2","is_open":true,"lib_rt":{"seats_total":20,"seats_used":1,"seats_booking":0}}
			]}}}}`)
			return
		}
		// Authorization URL endpoint
		http.SetCookie(w, &http.Cookie{Name: "SERVERID", Value: "s1"})
		http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: "tok"})
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	userWechatURL := "http://wechat.v2.traceint.com/index.php/graphql/?operationName=index&query=query%7BuserAuth%7BtongJi%7Brank%7D%7D%7D&code=081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP&state=1"
	res, err := svc.AuthenticateFromCookie(ctx, 42, do.AuthFromCookieRequest{Cookie: userWechatURL, Remember: true})
	require.NoError(t, err)
	require.True(t, res.Session.Authorized)
	require.Len(t, res.Libraries, 1)
	assert.Equal(t, 8, res.Libraries[0].LibraryID)
}

func TestGetAuthQRCode(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {})
	ctx := context.Background()

	qr, err := svc.GetAuthQRCode(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, qr)
	assert.NotEmpty(t, qr.AuthURL)
	assert.True(t, strings.HasPrefix(qr.ImageDataURL, "data:image/png;base64,"))
}

func TestGetCheckInAuthQRCode(t *testing.T) {
	svc := setupService(t, func(w http.ResponseWriter, r *http.Request) {})
	ctx := context.Background()

	qr, err := svc.GetCheckInAuthQRCode(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, qr)
	assert.NotEmpty(t, qr.AuthURL)
	assert.True(t, strings.HasPrefix(qr.ImageDataURL, "data:image/png;base64,"))
}
