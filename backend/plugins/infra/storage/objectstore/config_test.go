// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import "testing"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantError bool
	}{
		{
			name: "valid local config",
			cfg: Config{
				Driver: DriverLocal,
				Local:  LocalConfig{Root: "/data/storage"},
			},
			wantError: false,
		},
		{
			name: "invalid local config empty root",
			cfg: Config{
				Driver: DriverLocal,
				Local:  LocalConfig{Root: "  "},
			},
			wantError: true,
		},
		{
			name: "valid s3 config",
			cfg: Config{
				Driver: DriverS3,
				S3: ObjectConfig{
					Region:          "us-east-1",
					Bucket:          "my-bucket",
					AccessKeyID:     "akid",
					SecretAccessKey: "secret",
				},
			},
			wantError: false,
		},
		{
			name: "invalid s3 missing secret",
			cfg: Config{
				Driver: DriverS3,
				S3: ObjectConfig{
					Region:      "us-east-1",
					Bucket:      "my-bucket",
					AccessKeyID: "akid",
				},
			},
			wantError: true,
		},
		{
			name: "valid minio config",
			cfg: Config{
				Driver: DriverMinIO,
				MinIO: ObjectConfig{
					Endpoint:        "http://minio:9000",
					Region:          "us-east-1",
					Bucket:          "minio-bucket",
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
				},
			},
			wantError: false,
		},
		{
			name: "invalid minio missing endpoint",
			cfg: Config{
				Driver: DriverMinIO,
				MinIO: ObjectConfig{
					Region:          "us-east-1",
					Bucket:          "minio-bucket",
					AccessKeyID:     "minioadmin",
					SecretAccessKey: "minioadmin",
				},
			},
			wantError: true,
		},
		{
			name: "valid webdav config",
			cfg: Config{
				Driver: DriverWebDAV,
				WebDAV: WebDAVConfig{Endpoint: "https://dav.example.com"},
			},
			wantError: false,
		},
		{
			name: "invalid webdav missing endpoint",
			cfg: Config{
				Driver: DriverWebDAV,
				WebDAV: WebDAVConfig{Endpoint: "   "},
			},
			wantError: true,
		},
		{
			name: "unsupported driver",
			cfg: Config{
				Driver: Driver("ftp"),
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(tc.cfg)
			if (err != nil) != tc.wantError {
				t.Fatalf("ValidateConfig() err = %v, wantError = %v", err, tc.wantError)
			}
		})
	}
}

func TestMaskAndMergeSecrets(t *testing.T) {
	orig := Config{
		Driver: DriverS3,
		S3: ObjectConfig{
			Region:          "us-east-1",
			Bucket:          "my-bucket",
			AccessKeyID:     "real-access-key",
			SecretAccessKey: "real-secret-key",
		},
		WebDAV: WebDAVConfig{
			Endpoint: "https://dav.example.com",
			Password: "real-password",
		},
	}

	masked := MaskSecrets(orig)
	if masked.S3.AccessKeyID != ConfigMask {
		t.Errorf("expected S3.AccessKeyID to be masked, got %s", masked.S3.AccessKeyID)
	}
	if masked.S3.SecretAccessKey != ConfigMask {
		t.Errorf("expected S3.SecretAccessKey to be masked, got %s", masked.S3.SecretAccessKey)
	}
	if masked.WebDAV.Password != ConfigMask {
		t.Errorf("expected WebDAV.Password to be masked, got %s", masked.WebDAV.Password)
	}

	// Now pretend user updated some non-secret property while keeping mask
	next := masked
	next.S3.Region = "us-west-2"

	merged := MergeMaskedSecrets(next, orig)
	if merged.S3.AccessKeyID != "real-access-key" {
		t.Errorf("expected merged S3.AccessKeyID to be real-access-key, got %s", merged.S3.AccessKeyID)
	}
	if merged.S3.SecretAccessKey != "real-secret-key" {
		t.Errorf("expected merged S3.SecretAccessKey to be real-secret-key, got %s", merged.S3.SecretAccessKey)
	}
	if merged.WebDAV.Password != "real-password" {
		t.Errorf("expected merged WebDAV.Password to be real-password, got %s", merged.WebDAV.Password)
	}
	if merged.S3.Region != "us-west-2" {
		t.Errorf("expected merged S3.Region to be us-west-2, got %s", merged.S3.Region)
	}
}
