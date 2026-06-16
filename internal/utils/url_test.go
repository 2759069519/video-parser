package utils

import (
	"testing"
)

func TestResourcePlatform(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantPlat string
		wantErr  bool
	}{
		{"xiaohongshu main", "https://www.xiaohongshu.com/explore/123", "xhs", false},
		{"xiaohongshu link", "https://xhslink.com/abc", "xhs", false},
		{"xiaohongshu cdn", "https://sns-webpic-qc.xhscdn.com/123", "xhs", false},
		{"douyin main", "https://www.douyin.com/video/123", "douyin", false},
		{"douyin ies", "https://www.iesdouyin.com/share/video/123", "douyin", false},
		{"kuaishou main", "https://www.kuaishou.com/short-video/123", "kuaishou", false},
		{"kuaishou gifshow", "https://www.gifshow.com/video/123", "kuaishou", false},
		{"miyoushe", "https://www.miyoushe.com/ys/article/123", "miyoushe", false},
		{"doubao", "https://www.doubao.com/chat/123", "doubao", false},
		{"invalid scheme", "ftp://example.com/file", "", true},
		{"empty host", "https://", "", true},
		{"unknown domain", "https://www.google.com/video/123", "", true},
		{"http allowed", "http://www.douyin.com/video/123", "douyin", false},
		{"subdomain xhs", "https://img.xiaohongshu.com/123", "xhs", false},
		{"subdomain douyin", "https://v.douyin.com/abc", "douyin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plat, err := ResourcePlatform(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourcePlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if plat != tt.wantPlat {
				t.Errorf("ResourcePlatform() = %v, want %v", plat, tt.wantPlat)
			}
		})
	}
}

func TestIsAllowedDownloadURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"allowed douyin", "https://www.douyin.com/video/123", false},
		{"allowed xhs", "https://www.xiaohongshu.com/explore/123", false},
		{"disallowed google", "https://www.google.com", true},
		{"disallowed http example", "http://example.com/file.mp4", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsAllowedDownloadURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsAllowedDownloadURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectPlatformByHostname(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"douyin", "https://www.douyin.com/video/123", "douyin"},
		{"douyin subdomain", "https://v.douyin.com/abc", "douyin"},
		{"kuaishou", "https://www.kuaishou.com/video/123", "kuaishou"},
		{"kuaishou gifshow", "https://www.gifshow.com/video/123", "kuaishou"},
		{"xiaohongshu", "https://www.xiaohongshu.com/explore/123", "xhs"},
		{"xiaohongshu link", "https://xhslink.com/abc", "xhs"},
		{"miyoushe", "https://www.miyoushe.com/ys/article/123", "miyoushe"},
		{"unknown", "https://www.google.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectPlatformByHostname(tt.url); got != tt.want {
				t.Errorf("DetectPlatformByHostname() = %v, want %v", got, tt.want)
			}
		})
	}
}
