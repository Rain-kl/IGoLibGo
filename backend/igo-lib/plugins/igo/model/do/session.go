// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

// SessionResponse is the TraceInt cookie session for the current Wavelet user.
type SessionResponse struct {
	Authorized     bool   `json:"authorized"`
	Source         string `json:"source,omitempty"`
	SavedAt        string `json:"saved_at,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	CanAutoRestore bool   `json:"can_auto_restore"`
	CookieMasked   string `json:"cookie_masked,omitempty"`
}

// AuthFromCodeRequest authenticates a TraceInt session from a WeChat auth link/code.
type AuthFromCodeRequest struct {
	Code     string `json:"code" binding:"required"`
	Remember bool   `json:"remember"`
}

// AuthFromCookieRequest authenticates a TraceInt session from a raw cookie.
type AuthFromCookieRequest struct {
	Cookie   string `json:"cookie" binding:"required"`
	Remember bool   `json:"remember"`
}

// RefreshCookieRequest optionally supplies a new WeChat code to replace the cookie.
type RefreshCookieRequest struct {
	Code     string `json:"code"`
	Remember bool   `json:"remember"`
}

// SessionWorkflowResponse is returned after login/restore/refresh.
type SessionWorkflowResponse struct {
	Session   SessionResponse  `json:"session"`
	Libraries []LibrarySummary `json:"libraries,omitempty"`
	Message   string           `json:"message,omitempty"`
}
