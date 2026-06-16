package parser

import (
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"douyin main", "https://www.douyin.com/video/123", "douyin"},
		{"douyin subdomain", "https://v.douyin.com/abc", "douyin"},
		{"douyin ies", "https://www.iesdouyin.com/share/video/123", "douyin"},
		{"kuaishou main", "https://www.kuaishou.com/short-video/123", "kuaishou"},
		{"kuaishou gifshow", "https://www.gifshow.com/video/123", "kuaishou"},
		{"kuaishou chenzhongtech", "https://www.chenzhongtech.com/video/123", "kuaishou"},
		{"xiaohongshu", "https://www.xiaohongshu.com/explore/123", "xhs"},
		{"xiaohongshu link", "https://xhslink.com/abc", "xhs"},
		{"miyoushe", "https://www.miyoushe.com/ys/article/123", "miyoushe"},
		{"doubao", "https://www.doubao.com/chat/123", "doubao"},
		{"doubao subdomain", "https://api.doubao.com/chat/123", "doubao"},
		{"unknown", "https://www.google.com", ""},
		{"empty url", "", ""},
		{"invalid url", "://invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectPlatform(tt.url); got != tt.want {
				t.Errorf("DetectPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCleanURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple url", "https://www.douyin.com/video/123", "https://www.douyin.com/video/123"},
		{"url with space", "https://www.douyin.com/video/123 more text", "https://www.douyin.com/video/123"},
		{"url with newline", "https://www.douyin.com/video/123\n", "https://www.douyin.com/video/123"},
		{"url with trailing period", "https://www.douyin.com/video/123.", "https://www.douyin.com/video/123"},
		{"no url", "just text", "just text"},
		{"mixed case", "HTTP://WWW.DOUYIN.COM/VIDEO/123", "HTTP://WWW.DOUYIN.COM/VIDEO/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractCleanURL(tt.input); got != tt.want {
				t.Errorf("ExtractCleanURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
