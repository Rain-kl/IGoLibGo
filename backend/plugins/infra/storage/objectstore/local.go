// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type localBackend struct {
	root string
}

func newLocalBackend(cfg LocalConfig) (*localBackend, error) {
	root := filepath.Clean(cfg.Root)
	if root == "" {
		return nil, errors.New("local root is required")
	}
	return &localBackend{root: root}, nil
}

func (b *localBackend) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) (PutResult, error) {
	path, err := b.path(key)
	if err != nil {
		return PutResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(path), storageDirPerm); err != nil {
		return PutResult{}, fmt.Errorf("failed to create directory for %s: %w", path, err)
	}
	file, err := os.OpenFile( //nolint:gosec // path is constrained to the configured storage root.
		path,
		os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
		storageFilePerm,
	)
	if err != nil {
		return PutResult{}, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	if _, err := io.Copy(file, body); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return PutResult{}, fmt.Errorf("failed to write content to %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return PutResult{}, fmt.Errorf("failed to close file %s: %w", path, err)
	}
	return PutResult{Key: filepath.ToSlash(key)}, nil
}

func (b *localBackend) Get(_ context.Context, key string) (*Object, error) {
	path, err := b.path(key)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path) //nolint:gosec // path is constrained to the configured storage root.
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("failed to stat file %s: %w", path, err)
	}
	contentType := mime.TypeByExtension(filepath.Ext(path))
	if contentType == "" {
		contentType = defaultContentType
	}
	return &Object{Body: file, ContentLength: info.Size(), ContentType: contentType}, nil
}

func (b *localBackend) Delete(_ context.Context, key string) error {
	path, err := b.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	return nil
}

func (b *localBackend) Test(_ context.Context) error {
	return os.MkdirAll(b.root, storageDirPerm)
}

func (b *localBackend) path(key string) (string, error) {
	if filepath.IsAbs(key) {
		cleanPath := filepath.Clean(key)
		absRoot, err := filepath.Abs(b.root)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute root path: %w", err)
		}
		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute file path: %w", err)
		}
		rel, err := filepath.Rel(absRoot, absPath)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", errors.New("storage key escapes local root")
		}
		return cleanPath, nil
	}
	cleanKey := filepath.Clean(filepath.FromSlash(strings.TrimPrefix(key, "/")))
	if cleanKey == "." || cleanKey == "" || strings.HasPrefix(cleanKey, "..") {
		return "", fmt.Errorf("invalid local storage key %q", key)
	}
	path := filepath.Join(b.root, cleanKey)
	rel, err := filepath.Rel(b.root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("storage key escapes local root")
	}
	return path, nil
}
