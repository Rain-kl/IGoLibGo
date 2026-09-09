// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
)

func TestUserDTOJSON(t *testing.T) {
	user := UserDTO{
		ID:        999,
		Username:  "alice",
		Nickname:  "Alice Wonderland",
		Email:     "alice@example.com",
		AvatarURL: "https://example.com/avatar.png",
		IsActive:  true,
		IsAdmin:   true,
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("json.Marshal(user) error: %v", err)
	}

	var parsed UserDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.ID != user.ID {
		t.Errorf("got ID = %d, want %d", parsed.ID, user.ID)
	}
	if parsed.Username != user.Username {
		t.Errorf("got Username = %q, want %q", parsed.Username, user.Username)
	}
	if !parsed.IsAdmin {
		t.Errorf("expected IsAdmin to be true")
	}
}

func TestCreateUserRequestJSON(t *testing.T) {
	req := CreateUserRequest{
		Username: "bob",
		Password: "password123",
		Nickname: "Bob The Builder",
		Email:    "bob@example.com",
		IsAdmin:  false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal(req) error: %v", err)
	}

	var parsed CreateUserRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Username != req.Username || parsed.Email != req.Email {
		t.Errorf("parsed request mismatch: %+v", parsed)
	}
}
