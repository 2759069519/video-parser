package handler

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal", "video.mp4", "video.mp4"},
		{"with newline", "video\n.mp4", "video.mp4"},
		{"with quotes", "video\"test.mp4", "videotest.mp4"},
		{"with path", "path/to/video.mp4", "pathtovideo.mp4"},
		{"with backslash", "path\\to\\video.mp4", "pathtovideo.mp4"},
		{"empty", "", "file"},
		{"only spaces", "   ", "file"},
		{"with null", "video\x00.mp4", "video.mp4"},
		{"with single quote", "video'test.mp4", "videotest.mp4"},
		{"with carriage return", "video\r.mp4", "video.mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.input); got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAllowedImageType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{"jpeg", "image/jpeg", true},
		{"jpeg with params", "image/jpeg; charset=utf-8", true},
		{"png", "image/png", true},
		{"webp", "image/webp", true},
		{"avif", "image/avif", true},
		{"gif", "image/gif", false},
		{"svg", "image/svg+xml", false},
		{"html", "text/html", false},
		{"empty", "", false},
		{"uppercase", "IMAGE/JPEG", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedImageType(tt.contentType); got != tt.want {
				t.Errorf("isAllowedImageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"douyin url", "https://www.douyin.com/video/123", "https://www.douyin.com/video/123"},
		{"xhs url", "https://www.xiaohongshu.com/explore/123", "https://www.xiaohongshu.com/explore/123"},
		{"text with url", "请解析 https://www.douyin.com/video/123 谢谢", "https://www.douyin.com/video/123"},
		{"no url", "just text", "just text"},
		{"kuaishou", "https://www.kuaishou.com/short-video/123", "https://www.kuaishou.com/short-video/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractURL(tt.input); got != tt.want {
				t.Errorf("extractURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
