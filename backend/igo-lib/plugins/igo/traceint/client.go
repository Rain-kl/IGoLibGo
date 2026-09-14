// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package traceint talks to the TraceInt WeChat library APIs.
package traceint

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const (
	desktopUA  = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36 NetType/WIFI MicroMessenger/7.0.20.1781(0x6700143B) WindowsWechat(0x63070626)"
	tomorrowUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 NetType/WIFI MicroMessenger/7.0.20.1781(0x6700143B) WindowsWechat(0x63090719) XWEB/8391 Flue"
	appVersion = "2.0.11"
)

// Client is a TraceInt HTTP/GraphQL client.
type Client struct {
	HTTP       *http.Client
	MaxRetries int
}

func (c *Client) http() *http.Client {
	if c != nil && c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

func (c *Client) retries() int {
	if c != nil && c.MaxRetries > 0 {
		return c.MaxRetries
	}
	return 3
}

// GetCookie exchanges a WeChat code for a TraceInt cookie header.
func (c *Client) GetCookie(ctx context.Context, templates do.ProtocolTemplatesResponse, code string) (string, error) {
	reqURL := BuildAuthorizationURL(templates.GetCookieURLTemplate, code, templates.CookieAuthorizationReturnURL)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	cli := *c.http()
	cli.Jar = jar
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", desktopUA)
	resp, err := cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取 Cookie 失败: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.Request == nil || resp.Request.URL == nil {
		return "", errf("获取 Cookie 失败：空响应")
	}
	cookies := jar.Cookies(resp.Request.URL)
	if len(cookies) == 0 {
		cookies = resp.Cookies()
	}
	return BuildCookieHeader(cookies)
}

func (c *Client) graphql(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie, payload string, tomorrow bool) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, templates.GraphQLEndpointURL, bytes.NewReader([]byte(payload)))
	if err != nil {
		return nil, err
	}
	ua, referer, origin, ver := desktopUA, templates.GraphQLDefaultRefererURL, templates.GraphQLDefaultOriginURL, appVersion
	if tomorrow {
		ua, referer, origin, ver = tomorrowUA, templates.GraphQLTomorrowRefererURL, templates.GraphQLTomorrowOriginURL, "2.2.5"
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("App-Version", ver)
	req.Header.Set("app-version", ver)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Content-Type", "application/json")
	var lastErr error
	attempts := c.retries()
	for i := 0; i < attempts; i++ {
		if i > 0 {
			req, err = http.NewRequestWithContext(ctx, http.MethodPost, templates.GraphQLEndpointURL, bytes.NewReader([]byte(payload)))
			if err != nil {
				return nil, err
			}
			req.Header.Set("Cookie", cookie)
			req.Header.Set("Origin", origin)
			req.Header.Set("Referer", referer)
			req.Header.Set("User-Agent", ua)
			req.Header.Set("App-Version", ver)
			req.Header.Set("app-version", ver)
			req.Header.Set("Accept", "*/*")
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.http().Do(req)
		if err != nil {
			lastErr = fmt.Errorf("TraceInt 请求失败: %w", err)
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 500 && i+1 < attempts {
			lastErr = fmt.Errorf("TraceInt HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
			continue
		}
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("TraceInt HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
		}
		return body, nil
	}
	return nil, lastErr
}

// ListLibraries loads venues for a cookie.
func (c *Client) ListLibraries(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string) ([]do.LibrarySummary, error) {
	raw, err := c.graphql(ctx, templates, cookie, templates.QueryLibrariesTemplate, false)
	if err != nil {
		return nil, err
	}
	return mapLibraries(raw)
}

// GetLayout loads a venue seat map.
func (c *Client) GetLayout(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string, libraryID int) (*do.LibraryLayoutResponse, error) {
	raw, err := c.graphql(ctx, templates, cookie, FillLibID(templates.QueryLibraryLayoutTemplate, libraryID), false)
	if err != nil {
		return nil, err
	}
	return mapLayout(raw)
}

// GetRule loads venue booking rules.
func (c *Client) GetRule(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string, libraryID int) (*do.LibraryRuleResponse, error) {
	raw, err := c.graphql(ctx, templates, cookie, FillLibID(templates.QueryLibraryRuleTemplate, libraryID), false)
	if err != nil {
		return nil, err
	}
	return mapRule(raw, libraryID)
}

// GetReservation loads the current reservation.
func (c *Client) GetReservation(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string) (*do.ReservationResponse, error) {
	raw, err := c.graphql(ctx, templates, cookie, templates.QueryReservationInfoTemplate, false)
	if err != nil {
		return nil, err
	}
	return mapReservation(raw)
}

// ReserveSeat books a seat.
func (c *Client) ReserveSeat(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string, libraryID int, seatKey string) (bool, error) {
	raw, err := c.graphql(ctx, templates, cookie, FillSeat(templates.ReserveSeatTemplate, seatKey, libraryID), false)
	if err != nil {
		return false, err
	}
	return mapReserveOK(raw)
}

// CancelReservation cancels by sToken.
func (c *Client) CancelReservation(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie, token string) (bool, error) {
	payload := strings.ReplaceAll(templates.CancelReservationTemplate, "ReplaceMe", token)
	raw, err := c.graphql(ctx, templates, cookie, payload, false)
	if err != nil {
		return false, err
	}
	return mapCancelOK(raw), nil
}

// WarmUpTomorrow hits the prereserve layout query.
func (c *Client) WarmUpTomorrow(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string, libraryID int) error {
	_, err := c.graphql(ctx, templates, cookie, FillLibID(templates.TomorrowReservationWarmUpTemplate, libraryID), true)
	return err
}

// SaveTomorrow submits a tomorrow reservation.
// Original TraceInt payload appends a trailing "." to the seat key.
func (c *Client) SaveTomorrow(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie string, libraryID int, seatKey string) error {
	raw, err := c.graphql(ctx, templates, cookie, FillSeat(templates.TomorrowReservationSaveTemplate, TomorrowSeatKey(seatKey), libraryID), true)
	if err != nil {
		return err
	}
	_, err = parseRoot(raw)
	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
