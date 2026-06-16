package parser

import (
	"os"
	"testing"
)

func TestExtractMiyoushePostID(t *testing.T) {
	cases := map[string]string{
		"https://m.miyoushe.com/ys?channel=vivo/#/article/75779898":    "75779898",
		"https://www.miyoushe.com/ys/article/10475874":                 "10475874",
		"https://bbs-api.miyoushe.com/post/wapi/getPostFull?post_id=1": "1",
	}

	for rawURL, want := range cases {
		if got := extractMiyoushePostID(rawURL); got != want {
			t.Fatalf("extractMiyoushePostID(%q) = %q, want %q", rawURL, got, want)
		}
	}
}

func TestMiyousheParserLive(t *testing.T) {
	if os.Getenv("MIYOUSHE_LIVE_TEST") != "1" {
		t.Skip("set MIYOUSHE_LIVE_TEST=1 to run live miyoushe parser checks")
	}

	parser := NewMiyousheParser()

	video, contentType, err := parser.ParseURL("https://m.miyoushe.com/ys?channel=vivo/#/article/75779898")
	if err != nil {
		t.Fatalf("parse video: %v", err)
	}
	if contentType != "video" {
		t.Fatalf("video content type = %q, want video", contentType)
	}
	if result, ok := video.(*VideoResult); !ok || result.VideoURL == "" || result.CoverURL == "" {
		t.Fatalf("invalid video result: %#v", video)
	}

	atlas, contentType, err := parser.ParseURL("https://m.miyoushe.com/ys?channel=vivo/#/article/10475874")
	if err != nil {
		t.Fatalf("parse atlas: %v", err)
	}
	if contentType != "atlas" {
		t.Fatalf("atlas content type = %q, want atlas", contentType)
	}
	if result, ok := atlas.(*AtlasResult); !ok || len(result.Images) == 0 {
		t.Fatalf("invalid atlas result: %#v", atlas)
	}
}
