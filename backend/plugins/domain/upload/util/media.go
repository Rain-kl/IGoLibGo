// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"Wavelet/plugins/domain/upload/shared"
	"slices"
	"strings"
)

// IsImageExtension reports whether ext is a common image format.
func IsImageExtension(ext string) bool {
	return slices.Contains([]string{"jpg", "jpeg", "png", "webp", "gif"}, ext)
}

// IsArchiveExtension reports whether ext is a common archive format.
func IsArchiveExtension(ext string) bool {
	return slices.Contains([]string{"zip", "rar", "7z", "tar", "gz", "tgz", "bz2", "xz"}, ext)
}

// IsDocumentExtension reports whether ext is a common document format.
func IsDocumentExtension(ext string) bool {
	return slices.Contains([]string{"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "md", "csv", "json", "yaml", "yml", "xml"}, ext)
}

// NormalizeImageQuality normalizes the requested image quality query parameter.
func NormalizeImageQuality(quality string) string {
	q := strings.TrimSpace(strings.ToLower(quality))
	switch q {
	case shared.ImageQualityLow, shared.ImageQualityMedium, shared.ImageQualityHigh:
		return q
	default:
		return shared.ImageQualityOrigin
	}
}
