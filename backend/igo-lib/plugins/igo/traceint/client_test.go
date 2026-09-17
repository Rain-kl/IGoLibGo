package traceint

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"Wavelet/igo-lib/plugins/igo/model/do"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type crossHostAuthTransport struct {
	failWithMessage bool
}

func (t *crossHostAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "wechat.v2.traceint.com" {
		if t.failWithMessage {
			header := make(http.Header)
			header.Set("Content-Type", "text/html;charset=utf-8")
			header.Add("Set-Cookie", "SERVERID=srv456; Path=/")
			body := `<html><body><div class="text">微信授权失败（code been used, rid: 6aa79bc5）！</div></body></html>`
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		}
		header := make(http.Header)
		header.Add("Set-Cookie", "Authorization=Bearer token123; Path=/")
		header.Add("Set-Cookie", "SERVERID=srv456; Path=/")
		header.Set("Location", "https://web.traceint.com/web/index.html")
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     header,
			Body:       http.NoBody,
			Request:    req,
		}, nil
	}
	if req.URL.Host == "web.traceint.com" {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       http.NoBody,
			Request:    req,
		}, nil
	}
	return http.DefaultTransport.RoundTrip(req)
}

type checkInMockTransport struct{}

func (checkInMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	header := make(http.Header)
	header.Add("Set-Cookie", "wechatSESS_ID=sess123; Path=/")
	header.Set("Location", "https://web.traceint.com/web/index.html")
	return &http.Response{
		StatusCode: http.StatusFound,
		Header:     header,
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

func TestGetCookie_RedirectCrossHost(t *testing.T) {
	client := &Client{
		HTTP: &http.Client{
			Transport: &crossHostAuthTransport{},
		},
	}
	templates := do.ProtocolTemplatesResponse{
		GetCookieURLTemplate:         "http://wechat.v2.traceint.com/index.php/urlNew/auth.html?r=ReplaceMeByReturnUrl&code=ReplaceMeByCode&state=1",
		CookieAuthorizationReturnURL: "https://web.traceint.com/web/index.html",
	}

	cookie, err := client.GetCookie(context.Background(), templates, "081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP")
	require.NoError(t, err)
	assert.Contains(t, cookie, "Authorization=Bearer token123")
	assert.Contains(t, cookie, "SERVERID=srv456")
}

func TestGetCookie_ExtractErrorMessageOnFailure(t *testing.T) {
	client := &Client{
		HTTP: &http.Client{
			Transport: &crossHostAuthTransport{failWithMessage: true},
		},
	}
	templates := do.ProtocolTemplatesResponse{
		GetCookieURLTemplate:         "http://wechat.v2.traceint.com/index.php/urlNew/auth.html?r=ReplaceMeByReturnUrl&code=ReplaceMeByCode&state=1",
		CookieAuthorizationReturnURL: "https://web.traceint.com/web/index.html",
	}

	_, err := client.GetCookie(context.Background(), templates, "081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "微信授权失败（code been used")
}

func TestExtractCode_UserURL(t *testing.T) {
	rawURL := "http://wechat.v2.traceint.com/index.php/graphql/?operationName=index&query=query%7BuserAuth%7BtongJi%7Brank%7D%7D%7D&code=081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP&state=1"
	code, ok := ExtractCode(rawURL)
	require.True(t, ok)
	assert.Equal(t, "081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP", code)
}

func TestListLibraries_UsesDesktopGraphQLFingerprint(t *testing.T) {
	var got *http.Request
	client := &Client{
		HTTP: &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				got = req.Clone(req.Context())
				if req.Body != nil {
					_, _ = io.Copy(io.Discard, req.Body)
				}
				body := `{"data":{"userAuth":{"reserve":{"libs":[{"lib_id":1,"lib_name":"馆","lib_floor":"1","is_open":true,"lib_rt":{"seats_total":1,"seats_used":0,"seats_booking":0}}]}}}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(body)),
					Request:    req,
					Proto:      "HTTP/1.1",
					ProtoMajor: 1,
					ProtoMinor: 1,
				}, nil
			}),
		},
	}
	libs, err := client.ListLibraries(context.Background(), DefaultTemplates(), "Authorization=tok")
	require.NoError(t, err)
	require.Len(t, libs, 1)
	require.NotNil(t, got)
	assert.Equal(t, 1, got.ProtoMajor)
	assert.Equal(t, 1, got.ProtoMinor)
	assert.Equal(t, "keep-alive", got.Header.Get("Connection"))
	assert.Equal(t, "https://web.traceint.com", got.Header.Get("Origin"))
	assert.Equal(t, "https://web.traceint.com/web/index.html", got.Header.Get("Referer"))
	assert.Equal(t, desktopUA, got.Header.Get("User-Agent"))
	assert.Equal(t, "2.0.11", got.Header.Get("App-Version"))
	assert.Equal(t, "*/*", got.Header.Get("Accept"))
	assert.Equal(t, "application/json", got.Header.Get("Content-Type"))
	assert.Equal(t, "gzip, deflate, br", got.Header.Get("Accept-Encoding"))
	assert.Equal(t, "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7", got.Header.Get("Accept-Language"))
	assert.Equal(t, "same-site", got.Header.Get("Sec-Fetch-Site"))
	assert.Equal(t, "cors", got.Header.Get("Sec-Fetch-Mode"))
	assert.Equal(t, "empty", got.Header.Get("Sec-Fetch-Dest"))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildCookieHeader_CaseInsensitive(t *testing.T) {
	cookies := []*http.Cookie{
		{Name: "authorization", Value: "Bearer token"},
		{Name: "serverid", Value: "srv"},
	}
	header, err := BuildCookieHeader(cookies)
	require.NoError(t, err)
	assert.Contains(t, header, "authorization=Bearer token")
	assert.Contains(t, header, "serverid=srv")
}

func TestExchangeCheckInCode_Redirect(t *testing.T) {
	client := &Client{
		HTTP: &http.Client{
			Transport: checkInMockTransport{},
		},
	}
	templates := do.ProtocolTemplatesResponse{
		RemoteCheckInAuthURLTemplate:        "http://wechat.v2.traceint.com/index.php/wxApp/wechatAuth.html?r=ReplaceMeByReturnUrl&code=ReplaceMeByCode&state=1",
		RemoteCheckInAuthorizationReturnURL: "https://web.traceint.com/web/index.html",
	}

	token, _, err := client.ExchangeCheckInCode(context.Background(), templates, "081D0KGa1dplpM0ZzzIa1ss0tZ0D0KGP")
	require.NoError(t, err)
	assert.Equal(t, "sess123", token)
}
