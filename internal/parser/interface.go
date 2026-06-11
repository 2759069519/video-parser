package parser

import (
	"net/url"
	"strings"
)

type Parser interface {
	ParseURL(rawURL string) (interface{}, string, error)
}

type VideoResult struct {
	Type         string `json:"type"`
	VideoID      string `json:"video_id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	VideoURL     string `json:"video_url"`
	DownloadURL  string `json:"download_url"`
	CoverURL     string `json:"cover_url"`
	Duration     int    `json:"duration"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	ViewCount    int64  `json:"view_count"`
}

type AtlasResult struct {
	Type         string  `json:"type"`
	AtlasID      string  `json:"atlas_id"`
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	AuthorID     string  `json:"author_id"`
	Images       []Image `json:"images"`
	LikeCount    int64   `json:"like_count"`
	CommentCount int64   `json:"comment_count"`
}

type ProfileResult struct {
	Type        string  `json:"type"`
	UserID      string  `json:"user_id"`
	UserName    string  `json:"user_name"`
	Avatar      string  `json:"avatar"`
	Description string  `json:"description"`
	FanCount    int64   `json:"fan_count"`
	FollowCount int64   `json:"follow_count"`
	PhotoCount  int64   `json:"photo_count"`
	Photos      []Photo `json:"photos"`
}

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Photo struct {
	PhotoID      string `json:"photo_id"`
	Type         string `json:"type"`
	Caption      string `json:"caption"`
	CoverURL     string `json:"cover_url"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	ViewCount    int64  `json:"view_count"`
	Duration     int    `json:"duration"`
}

func DetectPlatform(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())

	if host == "kuaishou.com" || strings.HasSuffix(host, ".kuaishou.com") ||
		host == "gifshow.com" || strings.HasSuffix(host, ".gifshow.com") ||
		host == "chenzhongtech.com" || strings.HasSuffix(host, ".chenzhongtech.com") {
		return "kuaishou"
	}
	if host == "douyin.com" || strings.HasSuffix(host, ".douyin.com") ||
		host == "iesdouyin.com" || strings.HasSuffix(host, ".iesdouyin.com") {
		return "douyin"
	}
	if host == "xiaohongshu.com" || strings.HasSuffix(host, ".xiaohongshu.com") ||
		host == "xhslink.com" || strings.HasSuffix(host, ".xhslink.com") {
		return "xhs"
	}
	if host == "miyoushe.com" || strings.HasSuffix(host, ".miyoushe.com") {
		return "miyoushe"
	}
	return ""
}

func ExtractCleanURL(input string) string {
	lower := strings.ToLower(input)
	idx := strings.Index(lower, "http")
	if idx == -1 {
		return input
	}

	rest := input[idx:]
	for i, c := range rest {
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' {
			rest = rest[:i]
			break
		}
	}

	rest = strings.TrimRight(rest, ".,;:!?\"'")

	if _, err := url.Parse(rest); err == nil {
		return rest
	}
	return input
}
