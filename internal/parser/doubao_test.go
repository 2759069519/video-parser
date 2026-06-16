package parser

import (
	"testing"
)

func TestExtractVidFromURL(t *testing.T) {
	p := NewDoubaoParser()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"with video_id", "https://www.doubao.com/video-sharing?video_id=abc123", "abc123"},
		{"without video_id", "https://www.doubao.com/video-sharing", ""},
		{"other param", "https://www.doubao.com/video-sharing?other=value", ""},
		{"invalid url", "://invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.extractVidFromURL(tt.url); got != tt.want {
				t.Errorf("extractVidFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectImages(t *testing.T) {
	p := NewDoubaoParser()

	creations := []doubaoCreation{
		{
			Type: 1,
			Image: &doubaoImage{
				ImageOriRaw: &doubaoImageOriRaw{
					URL:    "https://example.com/image1.jpg",
					Width:  1920,
					Height: 1080,
				},
			},
		},
		{
			Type: 2,
			Video: &doubaoVideo{Vid: "video123"},
		},
		{
			Type: 1,
			Image: &doubaoImage{
				ImageOriRaw: &doubaoImageOriRaw{
					URL:    "https://example.com/image2.jpg?foo=bar",
					Width:  1280,
					Height: 720,
				},
			},
		},
	}

	images := p.collectImages(creations)

	if len(images) != 2 {
		t.Errorf("collectImages() returned %d images, want 2", len(images))
		return
	}

	if images[0].URL != "https://example.com/image1.jpg" {
		t.Errorf("images[0].URL = %v, want https://example.com/image1.jpg", images[0].URL)
	}
	if images[0].Width != 1920 {
		t.Errorf("images[0].Width = %v, want 1920", images[0].Width)
	}
	if images[1].URL != "https://example.com/image2.jpg?foo=bar" {
		t.Errorf("images[1].URL = %v, want https://example.com/image2.jpg?foo=bar", images[1].URL)
	}
}

func TestBuildAtlasResult(t *testing.T) {
	p := NewDoubaoParser()

	data := &doubaoPageData{
		Data: struct {
			ShareInfo       doubaoShareInfo       `json:"share_info"`
			MessageSnapshot doubaoMessageSnapshot `json:"message_snapshot"`
		}{
			ShareInfo: doubaoShareInfo{
				ShareName: "Test Title",
				User: doubaoShareUser{
					NickName: "Test Author",
				},
			},
		},
	}

	images := []Image{
		{URL: "https://example.com/img1.jpg", Width: 1920, Height: 1080},
	}

	result, contentType, err := p.buildAtlasResult(data, images)
	if err != nil {
		t.Errorf("buildAtlasResult() error = %v", err)
		return
	}

	if contentType != "atlas" {
		t.Errorf("buildAtlasResult() contentType = %v, want atlas", contentType)
	}
	if result.Title != "Test Title" {
		t.Errorf("result.Title = %v, want Test Title", result.Title)
	}
	if result.Author != "Test Author" {
		t.Errorf("result.Author = %v, want Test Author", result.Author)
	}
	if len(result.Images) != 1 {
		t.Errorf("result.Images length = %v, want 1", len(result.Images))
	}
}

func TestBuildAtlasResultEmptyImages(t *testing.T) {
	p := NewDoubaoParser()

	_, _, err := p.buildAtlasResult(nil, []Image{})
	if err == nil {
		t.Error("buildAtlasResult() with empty images should return error")
	}
}
