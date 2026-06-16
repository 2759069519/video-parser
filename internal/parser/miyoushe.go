package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MiyousheParser struct {
	client *http.Client
}

func NewMiyousheParser() *MiyousheParser {
	return &MiyousheParser{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

const miyousheMobileUA = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36"

var miyousheArticleRe = regexp.MustCompile(`/article/(\d+)`)

func (p *MiyousheParser) ParseURL(rawURL string) (interface{}, string, error) {
	postID := extractMiyoushePostID(rawURL)
	if postID == "" {
		return nil, "", errors.New("无法提取米游社文章ID")
	}

	post, err := p.fetchPost(postID)
	if err != nil {
		return nil, "", err
	}

	if len(post.VodList) > 0 {
		result := parseMiyousheVideo(post)
		if result.VideoURL == "" {
			return nil, "", errors.New("未找到米游社视频地址")
		}
		return result, "video", nil
	}
	if len(post.ImageList) > 0 || len(post.Post.Images) > 0 {
		return parseMiyousheAtlas(post), "atlas", nil
	}

	return nil, "", errors.New("未找到米游社视频或图片资源")
}

func extractMiyoushePostID(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil {
		if postID := u.Query().Get("post_id"); postID != "" {
			return postID
		}
		if matches := miyousheArticleRe.FindStringSubmatch(u.Fragment); len(matches) == 2 {
			return matches[1]
		}
	}
	if matches := miyousheArticleRe.FindStringSubmatch(rawURL); len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func (p *MiyousheParser) fetchPost(postID string) (*miyoushePostFull, error) {
	apiURL := fmt.Sprintf("https://bbs-api.miyoushe.com/post/wapi/getPostFull?gids=2&post_id=%s", url.QueryEscape(postID))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", miyousheMobileUA)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Origin", "https://m.miyoushe.com")
	req.Header.Set("Referer", "https://m.miyoushe.com/ys?channel=vivo/")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("米游社接口状态码: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, err
	}

	var apiResp miyousheAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析米游社JSON失败: %w", err)
	}
	if apiResp.Retcode != 0 {
		return nil, fmt.Errorf("米游社接口错误: %s", apiResp.Message)
	}
	return &apiResp.Data.Post, nil
}

func parseMiyousheVideo(post *miyoushePostFull) *VideoResult {
	vod := post.VodList[0]
	res := bestMiyousheResolution(vod.Resolutions)
	if res.URL == "" {
		res = bestMiyousheResolution(vod.BackupResolutions)
	}

	return &VideoResult{
		Type:         "video",
		VideoID:      firstNonEmpty(vod.ID, post.Post.PostID),
		Title:        post.Post.Subject,
		Author:       post.User.Nickname,
		AuthorID:     post.User.UID,
		VideoURL:     res.URL,
		DownloadURL:  res.URL,
		CoverURL:     firstNonEmpty(vod.Cover, post.Post.Cover),
		Duration:     vod.Duration,
		LikeCount:    post.Stat.LikeNum,
		CommentCount: post.Stat.ReplyNum,
		ViewCount:    firstInt64(vod.ViewNum, post.Stat.ViewNum),
	}
}

func parseMiyousheAtlas(post *miyoushePostFull) *AtlasResult {
	images := make([]Image, 0, len(post.ImageList))
	for _, item := range post.ImageList {
		if item.URL == "" {
			continue
		}
		images = append(images, Image{URL: item.URL, Width: item.Width, Height: item.Height})
	}
	if len(images) == 0 {
		for _, imageURL := range post.Post.Images {
			if imageURL != "" {
				images = append(images, Image{URL: imageURL})
			}
		}
	}

	return &AtlasResult{
		Type:         "atlas",
		AtlasID:      post.Post.PostID,
		Title:        firstNonEmpty(post.Post.Subject, miyousheContentTitle(post.Post.Content)),
		Author:       post.User.Nickname,
		AuthorID:     post.User.UID,
		Images:       images,
		LikeCount:    post.Stat.LikeNum,
		CommentCount: post.Stat.ReplyNum,
	}
}

func bestMiyousheResolution(resolutions []miyousheResolution) miyousheResolution {
	var best miyousheResolution
	for _, res := range resolutions {
		if res.URL == "" {
			continue
		}
		if best.URL == "" || res.Bitrate > best.Bitrate || res.Height > best.Height {
			best = res
		}
	}
	return best
}

func miyousheContentTitle(content string) string {
	var parsed struct {
		Describe string `json:"describe"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Describe)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

type miyousheAPIResponse struct {
	Retcode int    `json:"retcode"`
	Message string `json:"message"`
	Data    struct {
		Post miyoushePostFull `json:"post"`
	} `json:"data"`
}

type miyoushePostFull struct {
	Post      miyoushePost    `json:"post"`
	User      miyousheUser    `json:"user"`
	Stat      miyousheStat    `json:"stat"`
	ImageList []miyousheImage `json:"image_list"`
	VodList   []miyousheVod   `json:"vod_list"`
}

type miyoushePost struct {
	PostID  string   `json:"post_id"`
	Subject string   `json:"subject"`
	Content string   `json:"content"`
	Cover   string   `json:"cover"`
	Images  []string `json:"images"`
}

type miyousheUser struct {
	UID      string `json:"uid"`
	Nickname string `json:"nickname"`
}

type miyousheStat struct {
	ViewNum  int64 `json:"view_num"`
	ReplyNum int64 `json:"reply_num"`
	LikeNum  int64 `json:"like_num"`
}

type miyousheImage struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type miyousheVod struct {
	ID                string               `json:"id"`
	Duration          int                  `json:"duration"`
	Cover             string               `json:"cover"`
	ViewNum           int64                `json:"view_num"`
	Resolutions       []miyousheResolution `json:"resolutions"`
	BackupResolutions []miyousheResolution `json:"backup_resolutions"`
}

type miyousheResolution struct {
	URL        string `json:"url"`
	Definition string `json:"definition"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	Bitrate    int64  `json:"bitrate"`
	Size       string `json:"size"`
}

func (r *miyousheResolution) UnmarshalJSON(data []byte) error {
	type alias miyousheResolution
	var raw struct {
		alias
		Size interface{} `json:"size"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = miyousheResolution(raw.alias)
	switch size := raw.Size.(type) {
	case string:
		r.Size = size
	case float64:
		r.Size = strconv.FormatInt(int64(size), 10)
	}
	return nil
}
