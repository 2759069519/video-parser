package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type DoubaoParser struct {
	client *http.Client
}

func NewDoubaoParser() *DoubaoParser {
	return &DoubaoParser{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

const doubaoWebUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
const doubaoWeChatUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 NetType/WIFI MicroMessenger/7.0.20.1781(0x6700143B) WindowsWechat(0x63090c33) XWEB/14315 Flue"

var doubaoVidRe = regexp.MustCompile(`{\\\&quot;vid\\\&quot;:\\\&quot;([^\\]+)\\\&quot;`)
var doubaoSSRDataRe = regexp.MustCompile(`data-script-src="modern-run-router-data-fn" data-fn-args="(.*?)" nonce="`)

type doubaoCreation struct {
	Type  int            `json:"type"`
	ID    string         `json:"id"`
	Image *doubaoImage   `json:"image"`
	Video *doubaoVideo   `json:"video"`
}

type doubaoImage struct {
	ImageOriRaw *doubaoImageOriRaw `json:"image_ori_raw"`
}

type doubaoImageOriRaw struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type doubaoVideo struct {
	Vid         string         `json:"vid"`
	Cover       interface{}    `json:"cover"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Duration    float64        `json:"duration"`
	VideoType   string         `json:"video_type"`
	DownloadURL string         `json:"download_url"`
	VideoModel  string         `json:"video_model"`
}

type doubaoContentBlock struct {
	ContentV2 string `json:"content_v2"`
}

type doubaoMessage struct {
	ContentBlock []doubaoContentBlock `json:"content_block"`
}

type doubaoMessageSnapshot struct {
	MessageList []doubaoMessage `json:"message_list"`
}

type doubaoShareUser struct {
	NickName string `json:"nick_name"`
	Image    struct {
		OriginURL string `json:"origin_url"`
	} `json:"image"`
}

type doubaoShareInfo struct {
	ShareID   string          `json:"share_id"`
	ShareName string          `json:"share_name"`
	User      doubaoShareUser `json:"user"`
}

type doubaoPageData struct {
	Data struct {
		ShareInfo       doubaoShareInfo        `json:"share_info"`
		MessageSnapshot doubaoMessageSnapshot  `json:"message_snapshot"`
	} `json:"data"`
}

type doubaoSSRData []doubaoPageData

type doubaoGetPlayInfoResp struct {
	Code int `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		MediaType         string `json:"media_type"`
		PosterURL         string `json:"poster_url"`
		OriginalMediaInfo struct {
			Meta struct {
				Height   string  `json:"height"`
				Width    string  `json:"width"`
				Format   string  `json:"format"`
				Duration float64 `json:"duration"`
				CodecType string `json:"codec_type"`
				Definition string `json:"definition"`
			} `json:"meta"`
			MainURL   string `json:"main_url"`
			BackupURL string `json:"backup_url"`
		} `json:"original_media_info"`
	} `json:"data"`
}

func (p *DoubaoParser) ParseURL(rawURL string) (interface{}, string, error) {
	if strings.Contains(rawURL, "/video-sharing") {
		vid := p.extractVidFromURL(rawURL)
		if vid == "" {
			return nil, "", errors.New("无法从链接提取video_id")
		}
		return p.parseVideo(nil, vid)
	}

	html, err := p.fetchHTML(rawURL)
	if err != nil {
		return nil, "", err
	}

	pageData, err := p.parseSSRData(html)
	if err != nil {
		return nil, "", err
	}

	creations := p.extractCreations(pageData)

	for _, c := range creations {
		switch c.Type {
		case 2:
			if c.Video != nil && c.Video.Vid != "" {
				return p.parseVideo(pageData, c.Video.Vid)
			}
		case 1:
			if c.Image != nil && c.Image.ImageOriRaw != nil && c.Image.ImageOriRaw.URL != "" {
				images := p.collectImages(creations)
				if len(images) > 0 {
					return p.buildAtlasResult(pageData, images)
				}
			}
		}
	}

	vids := doubaoVidRe.FindAllStringSubmatch(html, -1)
	if len(vids) > 0 {
		return p.parseVideo(pageData, vids[0][1])
	}

	return nil, "", errors.New("未找到豆包视频或图片资源")
}

func (p *DoubaoParser) fetchHTML(pageURL string) (string, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", doubaoWebUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (p *DoubaoParser) parseSSRData(html string) (*doubaoPageData, error) {
	m := doubaoSSRDataRe.FindStringSubmatch(html)
	if len(m) < 2 {
		return nil, errors.New("未找到豆包页面数据")
	}

	jsonStr := m[1]
	jsonStr = strings.ReplaceAll(jsonStr, "&quot;", "\"")
	jsonStr = strings.ReplaceAll(jsonStr, "&amp;", "&")

	var rawArr []interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawArr); err != nil {
		return nil, fmt.Errorf("解析豆包页面数据失败: %w", err)
	}

	for _, item := range rawArr {
		if obj, ok := item.(map[string]interface{}); ok {
			if dataField, ok := obj["data"]; ok {
				reencoded, err := json.Marshal(obj)
				if err != nil {
					continue
				}
				var pageData doubaoPageData
				if err := json.Unmarshal(reencoded, &pageData); err == nil && dataField != nil {
					if d, ok := dataField.(map[string]interface{}); ok {
						if ms, ok := d["message_snapshot"]; ok && ms != nil {
							return &pageData, nil
						}
					}
				}
			}
		}
	}
	return nil, errors.New("未找到有效的页面数据")
}

func (p *DoubaoParser) extractCreations(data *doubaoPageData) []doubaoCreation {
	var creations []doubaoCreation
	for _, msg := range data.Data.MessageSnapshot.MessageList {
		for _, block := range msg.ContentBlock {
			if block.ContentV2 == "" {
				continue
			}
			var contentV2 struct {
				CreationBlock *struct {
					Creations []doubaoCreation `json:"creations"`
				} `json:"creation_block"`
			}
			if err := json.Unmarshal([]byte(block.ContentV2), &contentV2); err != nil {
				continue
			}
			if contentV2.CreationBlock != nil {
				creations = append(creations, contentV2.CreationBlock.Creations...)
			}
		}
	}
	return creations
}

func (p *DoubaoParser) collectImages(creations []doubaoCreation) []Image {
	var images []Image
	for _, c := range creations {
		if c.Type == 1 && c.Image != nil && c.Image.ImageOriRaw != nil && c.Image.ImageOriRaw.URL != "" {
			rawURL := strings.ReplaceAll(c.Image.ImageOriRaw.URL, "&amp;", "&")
			images = append(images, Image{
				URL:    rawURL,
				Width:  c.Image.ImageOriRaw.Width,
				Height: c.Image.ImageOriRaw.Height,
			})
		}
	}
	return images
}

func (p *DoubaoParser) getShareName(data *doubaoPageData) string {
	if data == nil {
		return ""
	}
	return data.Data.ShareInfo.ShareName
}

func (p *DoubaoParser) getShareUser(data *doubaoPageData) (string, string) {
	if data == nil {
		return "", ""
	}
	u := data.Data.ShareInfo.User
	return u.NickName, u.Image.OriginURL
}

func (p *DoubaoParser) buildAtlasResult(data *doubaoPageData, images []Image) (*AtlasResult, string, error) {
	if len(images) == 0 {
		return nil, "", errors.New("未找到图片")
	}

	author, _ := p.getShareUser(data)
	result := &AtlasResult{
		Type:   "atlas",
		Title:  p.getShareName(data),
		Author: author,
		Images: images,
	}

	return result, "atlas", nil
}

func (p *DoubaoParser) parseVideo(data *doubaoPageData, vid string) (*VideoResult, string, error) {
	videoInfo, err := p.fetchVideoInfo(vid)
	if err != nil {
		return nil, "", err
	}

	author, _ := p.getShareUser(data)
	posterURL := videoInfo.Data.PosterURL
	meta := videoInfo.Data.OriginalMediaInfo.Meta
	mainURL := videoInfo.Data.OriginalMediaInfo.MainURL
	backupURL := videoInfo.Data.OriginalMediaInfo.BackupURL

	mainURL = strings.ReplaceAll(mainURL, "&download=true", "")
	mainURL = strings.TrimSuffix(mainURL, "?download=true")
	backupURL = strings.ReplaceAll(backupURL, "&download=true", "")
	backupURL = strings.TrimSuffix(backupURL, "?download=true")

	var width, height int
	fmt.Sscanf(meta.Width, "%d", &width)
	fmt.Sscanf(meta.Height, "%d", &height)

	result := &VideoResult{
		Type:        "video",
		VideoID:     vid,
		Title:       p.getShareName(data),
		Author:      author,
		VideoURL:    mainURL,
		DownloadURL: backupURL,
		CoverURL:    posterURL,
		Duration:    int(meta.Duration * 1000),
	}

	_ = width
	_ = height

	return result, "video", nil
}

func (p *DoubaoParser) fetchVideoInfo(vid string) (*doubaoGetPlayInfoResp, error) {
	apiURL := fmt.Sprintf(
		"https://www.doubao.com/samantha/media/get_play_info?version_code=20800&language=zh-CN&device_platform=web&aid=497858&real_aid=497858&pkg_type=release_version&pc_version=2.51.7&samantha_web=1&use-olympus-account=1",
	)

	body, err := json.Marshal(map[string]string{"key": vid})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", doubaoWeChatUA)
	req.Header.Set("Origin", "https://www.doubao.com")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求豆包视频API失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("豆包视频API状态码: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, err
	}

	var result doubaoGetPlayInfoResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析豆包视频响应失败: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("豆包视频API错误: %s", result.Msg)
	}

	if result.Data.OriginalMediaInfo.MainURL == "" {
		return nil, errors.New("未找到豆包无水印视频地址")
	}

	return &result, nil
}

func (p *DoubaoParser) extractVidFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	if vid := q.Get("video_id"); vid != "" {
		return vid
	}
	return ""
}
