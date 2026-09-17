// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

// Package traceint talks to the TraceInt WeChat library APIs.
package traceint

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"time"
)

const (
	desktopUA             = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36 NetType/WIFI MicroMessenger/7.0.20.1781(0x6700143B) WindowsWechat(0x63070626)"
	tomorrowUA            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36 NetType/WIFI MicroMessenger/7.0.20.1781(0x6700143B) WindowsWechat(0x63090719) XWEB/8391 Flue"
	appVersion            = "2.0.11"
	defaultRetries        = 3
	maxErrorSnippetLength = 200
)

// Client is a TraceInt HTTP/GraphQL client.
type Client struct {
	HTTP       *http.Client
	MaxRetries int
}

func (c *Client) http() *http.Client {
	base := &http.Client{Timeout: 15 * time.Second}
	if c != nil && c.HTTP != nil {
		cloned := *c.HTTP
		base = &cloned
	}
	base.Transport = forceHTTP1(base.Transport)
	return base
}

func forceHTTP1(rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	tr, ok := rt.(*http.Transport)
	if !ok {
		return rt
	}
	clone := tr.Clone()
	clone.ForceAttemptHTTP2 = false
	clone.TLSNextProto = map[string]func(authority string, c *tls.Conn) http.RoundTripper{}
	return clone
}

func (c *Client) retries() int {
	if c != nil && c.MaxRetries > 0 {
		return c.MaxRetries
	}
	return defaultRetries
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
	cli.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		// Stop at the redirect response from the auth endpoint so that Set-Cookie headers
		// are not lost due to cross-host redirect (e.g. wechat.v2.traceint.com -> web.traceint.com).
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", desktopUA)
	resp, err := cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("获取 Cookie 失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.Request == nil || resp.Request.URL == nil {
		return "", errf("获取 Cookie 失败：空响应")
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 && req.URL != nil {
		cookies = jar.Cookies(req.URL)
	}
	if len(cookies) == 0 {
		cookies = jar.Cookies(resp.Request.URL)
	}

	header, err := BuildCookieHeader(cookies)
	if err != nil {
		if msg := extractTraceIntErrorMessage(string(body)); msg != "" {
			return "", fmt.Errorf("获取 Cookie 失败: %s", msg)
		}
		return "", err
	}
	return header, nil
}

func applyDesktopGraphQLHeaders(req *http.Request, cookie, origin, referer, ua, ver string) {
	req.Proto = "HTTP/1.1"
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", origin)
	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", ua)
	req.Header.Set("App-Version", ver)
	req.Header.Set("app-version", ver)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Expect")
}

func decodeHTTPBody(encoding string, body []byte) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "identity":
		return body, nil
	case "gzip":
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		defer func() { _ = r.Close() }()
		return io.ReadAll(r)
	default:
		return body, nil
	}
}

func (c *Client) graphql(ctx context.Context, templates do.ProtocolTemplatesResponse, cookie, payload string, tomorrow bool) ([]byte, error) {
	ua, referer, origin, ver := desktopUA, templates.GraphQLDefaultRefererURL, templates.GraphQLDefaultOriginURL, appVersion
	if tomorrow {
		ua, referer, origin, ver = tomorrowUA, templates.GraphQLTomorrowRefererURL, templates.GraphQLTomorrowOriginURL, "2.2.5"
	}
	payloadBytes := []byte(payload)
	var lastErr error
	attempts := c.retries()
	cli := c.http()
	for i := 0; i < attempts; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, templates.GraphQLEndpointURL, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, err
		}
		applyDesktopGraphQLHeaders(req, cookie, origin, referer, ua, ver)
		resp, err := cli.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("TraceInt 请求失败: %w", err)
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		body, err = decodeHTTPBody(resp.Header.Get("Content-Encoding"), body)
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode >= http.StatusInternalServerError && i+1 < attempts {
			lastErr = fmt.Errorf("TraceInt HTTP %d: %s", resp.StatusCode, truncate(string(body), maxErrorSnippetLength))
			continue
		}
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("TraceInt HTTP %d: %s", resp.StatusCode, truncate(string(body), maxErrorSnippetLength))
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

var (
	traceintErrorDivRE   = regexp.MustCompile(`(?i)<div\s+class=["']text["'][^>]*>([^<]+)</div>`)
	traceintTitleErrorRE = regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
)

func extractTraceIntErrorMessage(html string) string {
	if m := traceintErrorDivRE.FindStringSubmatch(html); len(m) >= minRegexMatches {
		msg := strings.TrimSpace(m[1])
		if msg != "" {
			return msg
		}
	}
	if m := traceintTitleErrorRE.FindStringSubmatch(html); len(m) >= minRegexMatches {
		title := strings.TrimSpace(m[1])
		if strings.Contains(title, "错误") || strings.Contains(strings.ToLower(title), "error") {
			return title
		}
	}
	return ""
}
