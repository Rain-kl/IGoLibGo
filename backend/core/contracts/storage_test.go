// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package contracts

import (
	"encoding/json"
	"testing"
)

func TestStorageConfigDTOJSON(t *testing.T) {
	cfg := StorageConfigDTO{
		Driver: StorageDriverS3,
		Local: LocalStorageConfigDTO{
			Root: "/uploads",
		},
		S3: ObjectStorageConfigDTO{
			Endpoint:        "https://s3.amazonaws.com",
			Region:          "us-east-1",
			Bucket:          "my-bucket",
			AccessKeyID:     "AKIA...",
			SecretAccessKey: "Secret...",
			CDNURL:          "https://cdn.example.com",
		},
		WebDAV: WebDAVStorageConfigDTO{
			URL:      "https://dav.example.com",
			Username: "user",
			Password: "password",
			Root:     "/dav",
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("json.Marshal(cfg) error: %v", err)
	}

	var parsed StorageConfigDTO
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json.Unmarshal(data) error: %v", err)
	}

	if parsed.Driver != StorageDriverS3 {
		t.Errorf("got driver %q, want %q", parsed.Driver, StorageDriverS3)
	}
	if parsed.S3.Bucket != "my-bucket" {
		t.Errorf("got s3 bucket %q, want %q", parsed.S3.Bucket, "my-bucket")
	}
	if parsed.WebDAV.Root != "/dav" {
		t.Errorf("got webdav root %q, want %q", parsed.WebDAV.Root, "/dav")
	}
}
