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
