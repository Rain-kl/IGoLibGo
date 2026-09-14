// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package traceint

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	rawCodeLength    = 32
	minRegexMatches  = 2
	minCookieCount   = 2
	minJWTSegments   = 2
	minMaskLength    = 16
	maskPrefixLength = 12
)

var codeRE = regexp.MustCompile(`(?i)code=([A-Za-z0-9]{32})`)

// ExtractCode pulls a 32-char WeChat code from a URL or raw code string.
func ExtractCode(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if len(raw) == rawCodeLength {
		for _, r := range raw {
			ok := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
			if !ok {
				return "", false
			}
		}
		return raw, true
	}
	m := codeRE.FindStringSubmatch(raw)
	if len(m) != minRegexMatches {
		return "", false
	}
	return m[1], true
}

// BuildCookieHeader prefers Authorization then SERVERID, matching the desktop client.
func BuildCookieHeader(cookies []*http.Cookie) (string, error) {
	if len(cookies) < minCookieCount {
		return "", errf("Cookie不包含关键身份信息，可能是code过期，重新填写含code的链接")
	}
	var auth, server string
	for _, c := range cookies {
		pair := c.Name + "=" + c.Value
		switch c.Name {
		case "Authorization":
			auth = pair
		case "SERVERID":
			server = pair
		}
	}
	if auth != "" && server != "" {
		return auth + "; " + server, nil
	}
	return cookies[1].Name + "=" + cookies[1].Value + "; " + cookies[0].Name + "=" + cookies[0].Value, nil
}

// CookieExpiration parses expireAt/exp from the Authorization JWT.
func CookieExpiration(cookie string) *time.Time {
	token := authorizationToken(cookie)
	parts := strings.Split(token, ".")
	if len(parts) < minJWTSegments {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil
		}
	}
	var body map[string]any
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil
	}
	for _, key := range []string{"expireAt", "exp"} {
		if sec, ok := unixSeconds(body[key]); ok {
			t := time.Unix(sec, 0)
			return &t
		}
	}
	return nil
}

func authorizationToken(cookie string) string {
	for _, seg := range strings.Split(cookie, ";") {
		trimmed := strings.TrimSpace(seg)
		if strings.HasPrefix(strings.ToLower(trimmed), "authorization=") {
			return strings.TrimSpace(trimmed[len("Authorization="):])
		}
	}
	return strings.TrimSpace(cookie)
}

func unixSeconds(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	case string:
		var i int64
		for _, r := range n {
			if r < '0' || r > '9' {
				return 0, false
			}
			i = i*10 + int64(r-'0')
		}
		return i, true
	default:
		return 0, false
	}
}

// MaskCookie redacts a cookie for API responses.
func MaskCookie(cookie string) string {
	if len(cookie) <= minMaskLength {
		return "***"
	}
	return cookie[:maskPrefixLength] + "..."
}
