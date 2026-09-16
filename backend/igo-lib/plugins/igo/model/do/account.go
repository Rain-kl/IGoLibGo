// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package do

import "time"

// AccountDTO is the public account view without raw credentials.
type AccountDTO struct {
	ID               uint64     `json:"id,string"`
	Name             string     `json:"name"`
	HasCookie        bool       `json:"has_cookie"`
	CookieMasked     string     `json:"cookie_masked,omitempty"`
	CookieExpiresAt  *time.Time `json:"cookie_expires_at,omitempty"`
	HasCheckinToken  bool       `json:"has_checkin_token"`
	CheckinExpiresAt *time.Time `json:"checkin_expires_at,omitempty"`
	Nickname         string     `json:"nickname"`
	School           string     `json:"school"`
	StudentName      string     `json:"student_name"`
	StudentNumber    string     `json:"student_number"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// CreateAccountRequest creates an account. Cookie and check-in token are optional.
type CreateAccountRequest struct {
	Name         string `json:"name" binding:"required"`
	Cookie       string `json:"cookie"`
	CheckinToken string `json:"checkin_token"`
}

// UpdateAccountRequest updates display fields.
type UpdateAccountRequest struct {
	Name string `json:"name"`
}

// AccountLoginRequest stores occupy credentials from a WeChat link or cookie.
type AccountLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// AccountCheckinAuthRequest stores WeChat check-in credentials.
type AccountCheckinAuthRequest struct {
	Code string `json:"code" binding:"required"`
}

// AccountCheckinAuthResponse is returned after exchanging a check-in code.
type AccountCheckinAuthResponse struct {
	Account AccountDTO             `json:"account"`
	Device  *CheckInDeviceResponse `json:"device,omitempty"`
}
