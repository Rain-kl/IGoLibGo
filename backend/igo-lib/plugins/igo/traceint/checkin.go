// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package traceint

import (
	"Wavelet/igo-lib/plugins/igo/model/do"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const checkInPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA0dmmkW4xPa+HhBTyaa0d
gAb0fVZRS67jK4y15BQthjJ/ZuUZQmrbGqhG7rwnxfm7g+nFH9zEyRU5KLX3ty9j
pNrPjyg7FBF9OvBDYHEt83b77W3mfBjpmoTJOt27E7RZ4InHqJQjqSEo4bw1PDz2
OBmtlNIlXMu0VA8I0Bh39hBBnm0oouRV7FdqEzAp8nsF7a3VuBYpx9xek+cRVip0
pMXI1AXM6bmyWWNzV0oikQW4ZIbutgDziTMeW28zl/hRbW9Ht34w0sWYyxumuLr1
qweW3qnxycn3zn47weFYe6nJp71z+lgVtNTGtowNPPqBLXqusvwf+uNhSy1wKQFp
UwIDAQAB
-----END PUBLIC KEY-----`

const (
	checkInAuthUA = "Mozilla/5.0 (iPad; CPU OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.75(0x18004b21) NetType/WIFI Language/zh_CN miniProgram/wx3b9352e6b254ed2b"
	checkInAPIUA  = "Mozilla/5.0 (iPad; CPU OS 27_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.75(0x18004b21) NetType/WIFI Language/zh_CN"
)

// ExchangeCheckInCode exchanges a WeChat code for wechatSESS_ID.
func (c *Client) ExchangeCheckInCode(ctx context.Context, templates do.ProtocolTemplatesResponse, code string) (token string, expiresAt *time.Time, err error) {
	reqURL := BuildAuthorizationURL(templates.RemoteCheckInAuthURLTemplate, code, templates.RemoteCheckInAuthorizationReturnURL)
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", nil, err
	}
	cli := *c.http()
	cli.Jar = jar
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", checkInAuthUA)
	req.Header.Set("Referer", templates.RemoteCheckInAuthRefererURL)
	resp, err := cli.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("获取签到授权失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	token, expiresAt = extractWechatSess(resp)
	if token == "" {
		return "", nil, errf("授权响应未返回 wechatSESS_ID，授权链接可能已被使用或已过期")
	}
	return token, expiresAt, nil
}

// GetCheckInDevices loads profile + beacon UUIDs.
func (c *Client) GetCheckInDevices(ctx context.Context, templates do.ProtocolTemplatesResponse, token string) (*do.CheckInDeviceResponse, error) {
	form := url.Values{"t": {token}}.Encode()
	raw, err := c.checkInForm(ctx, templates.RemoteCheckInDevicesEndpointURL, templates.RemoteCheckInAPIRefererURL, form)
	if err != nil {
		return nil, err
	}
	return mapDevices(raw)
}

// GetCheckInServerTime returns TraceInt unix time as a string.
func (c *Client) GetCheckInServerTime(ctx context.Context, templates do.ProtocolTemplatesResponse) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, templates.RemoteCheckInTimeEndpointURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", checkInAPIUA)
	req.Header.Set("Referer", templates.RemoteCheckInAPIRefererURL)
	resp, err := c.http().Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

// SignCheckIn submits a beacon check-in.
func (c *Client) SignCheckIn(ctx context.Context, templates do.ProtocolTemplatesResponse, token string, req do.CheckInSignRequest, serverTime string) (*do.CheckInSignResponse, error) {
	devices, err := encodeJSONBase64([]any{[]any{strings.ToUpper(req.BeaconUUID), req.Major, req.Minor}})
	if err != nil {
		return nil, err
	}
	location, err := encodeJSONBase64([]any{req.Latitude, req.Longitude})
	if err != nil {
		return nil, err
	}
	pass, err := encryptTimestamp(serverTime)
	if err != nil {
		return nil, err
	}
	form := url.Values{
		"t":        {token},
		"devices":  {devices},
		"location": {location},
		"pass":     {pass},
	}.Encode()
	raw, err := c.checkInForm(ctx, templates.RemoteCheckInSignEndpointURL, templates.RemoteCheckInAPIRefererURL, form)
	if err != nil {
		return nil, err
	}
	return mapSign(raw)
}

func (c *Client) checkInForm(ctx context.Context, endpoint, referer, form string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", checkInAPIUA)
	req.Header.Set("Referer", referer)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("签到 HTTP %d: %s", resp.StatusCode, truncate(string(body), maxErrorSnippetLength))
	}
	return body, nil
}

func extractWechatSess(resp *http.Response) (string, *time.Time) {
	for _, c := range resp.Cookies() {
		if strings.EqualFold(c.Name, "wechatSESS_ID") && c.Value != "" {
			var exp *time.Time
			if !c.Expires.IsZero() {
				t := c.Expires.UTC()
				exp = &t
			}
			return c.Value, exp
		}
	}
	for _, h := range resp.Header.Values("Set-Cookie") {
		first, _, _ := strings.Cut(h, ";")
		name, val, ok := strings.Cut(first, "=")
		if ok && strings.EqualFold(strings.TrimSpace(name), "wechatSESS_ID") {
			return strings.TrimSpace(val), nil
		}
	}
	return "", nil
}

func mapDevices(raw []byte) (*do.CheckInDeviceResponse, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, errf("设备信息响应格式无效")
	}
	if root["code"] != nil && asInt(root["code"]) != 0 {
		return nil, errf(orDefault(asString(root["msg"]), "设备信息请求失败"))
	}
	data, _ := root["data"].(map[string]any)
	user, _ := data["user"].(map[string]any)
	var uuids []string
	if arr, ok := data["devices"].([]any); ok {
		for _, v := range arr {
			s := strings.TrimSpace(asString(v))
			if s != "" {
				uuids = append(uuids, strings.ToUpper(s))
			}
		}
	}
	return &do.CheckInDeviceResponse{
		Nickname:      asString(user["user_nick"]),
		School:        asString(user["user_sch"]),
		StudentName:   asString(user["user_student_name"]),
		StudentNumber: asString(user["user_student_no"]),
		BeaconUUIDs:   uuids,
	}, nil
}

func mapSign(raw []byte) (*do.CheckInSignResponse, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, errf("签到响应格式无效")
	}
	if root["code"] != nil && asInt(root["code"]) != 0 {
		return nil, errf(orDefault(asString(root["msg"]), "签到失败"))
	}
	msg := asString(root["msg"])
	data, _ := root["data"].(map[string]any)
	out := &do.CheckInSignResponse{Message: orDefault(msg, "验证成功")}
	if data != nil {
		if v, ok := data["status"]; ok {
			n := asInt(v)
			out.Status = &n
		}
		if v, ok := data["lib_id"]; ok {
			n := asInt(v)
			out.LibraryID = &n
		}
		out.LibraryName = asString(data["lib_name"])
		out.LibraryFloor = asString(data["lib_floor"])
		out.SeatKey = asString(data["seat_key"])
		out.SeatName = asString(data["seat_name"])
	}
	return out, nil
}

func encodeJSONBase64(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func encryptTimestamp(ts string) (string, error) {
	block, _ := pem.Decode([]byte(checkInPublicKeyPEM))
	if block == nil {
		return "", errf("签到公钥无效")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errf("签到公钥类型错误")
	}
	//nolint:staticcheck // TraceInt remote protocol requires PKCS#1 v1.5 RSA encryption
	cipher, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(ts))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipher), nil
}
