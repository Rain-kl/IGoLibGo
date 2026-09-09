// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
)

func TestAuthSourceDTOJSON(t *testing.T) {
	src := AuthSourceDTO{
		ID:                 10,
		Name:               "github",
		Type:               "oauth2",
		DisplayName:        "GitHub OAuth",
		ClientID:           "client-123",
		ClientSecret:       "secret-456",
		OpenIDDiscoveryURL: "https://github.com/login/oauth/authorize",
		Scopes:             "read:user,user:email",
		IconURL:            "https://github.com/icon.png",
		IsActive:           true,
	}

	data, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("json.Marshal(src) error: %v", err)
	}

	var parsed AuthSourceDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.ID != src.ID || parsed.Name != src.Name {
		t.Errorf("parsed AuthSourceDTO mismatch: %+v", parsed)
	}
}

func TestAuthSourceViewDTOJSON(t *testing.T) {
	view := AuthSourceViewDTO{
		ID:                     10,
		Name:                   "github",
		Type:                   "oauth2",
		DisplayName:            "GitHub OAuth",
		IsActive:               true,
		IconURL:                "https://github.com/icon.png",
		ClientSecretConfigured: true,
	}

	data, err := json.Marshal(view)
	if err != nil {
		t.Fatalf("json.Marshal(view) error: %v", err)
	}

	var parsed AuthSourceViewDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Name != view.Name || !parsed.ClientSecretConfigured {
		t.Errorf("parsed AuthSourceViewDTO mismatch: %+v", parsed)
	}
}

func TestOAuthUserInfoDTOJSON(t *testing.T) {
	info := OAuthUserInfoDTO{
		ID:                123,
		Sub:               "gh-12345",
		Username:          "octocat",
		PreferredUsername: "octocat_pref",
		Email:             "octo@github.com",
		Name:              "The Octocat",
		Active:            true,
		AvatarURL:         "https://avatars.githubusercontent.com/u/583231",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal(info) error: %v", err)
	}

	var parsed OAuthUserInfoDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Sub != info.Sub || parsed.Email != info.Email {
		t.Errorf("parsed OAuthUserInfoDTO mismatch: %+v", parsed)
	}
}
