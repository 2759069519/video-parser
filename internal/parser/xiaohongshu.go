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

type XiaohongshuParser struct {
	client *http.Client
}

func NewXiaohongshuParser() *XiaohongshuParser {
	return &XiaohongshuParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

const xhsAndroidUA = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36"
const xhsPCUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"

func (p *XiaohongshuParser) ParseURL(rawURL string) (interface{}, string, error) {
	fullURL, err := p.resolveURL(rawURL)
	if err != nil {
		return nil, "", err
	}

	html, err := p.fetchHTML(fullURL)
	if err != nil {
		return nil, "", err
	}

	state, err := p.extractInitialState(html)
	if err != nil {
		return nil, "", err
	}

	note, err := p.findNote(state)
	if err != nil {
		return nil, "", err
	}

	noteType, _ := note["type"].(string)
	if noteType == "" {
		if _, hasVideo := note["video"]; hasVideo {
			noteType = "video"
		} else if _, hasImages := note["imageList"]; hasImages {
			noteType = "normal"
		}
	}

	switch noteType {
	case "video":
		return p.parseVideo(note)
	case "normal", "atlas":
		return p.parseAtlas(note)
	default:
		if _, hasVideo := note["video"]; hasVideo {
			return p.parseVideo(note)
		}
		if _, hasImages := note["imageList"]; hasImages {
			return p.parseAtlas(note)
		}
		return nil, "", fmt.Errorf("不支持的小红书笔记类型: %s (或内容为空，需要在小红书APP内查看)", noteType)
	}
}

func (p *XiaohongshuParser) resolveURL(rawURL string) (string, error) {
	if strings.Contains(rawURL, "xhslink.com") {
		client := &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		req, err := http.NewRequest("GET", rawURL, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", xhsAndroidUA)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")

		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			location := resp.Header.Get("Location")
			if location != "" {
				return location, nil
			}
		}
		return "", fmt.Errorf("无法获取短链重定向，状态码: %d", resp.StatusCode)
	}

	if !strings.Contains(rawURL, "/explore/") {
		if u := cleanXhsShareURL(rawURL); u != "" {
			return u, nil
		}
	}
	return rawURL, nil
}

func cleanXhsShareURL(rawURL string) string {
	if !strings.Contains(rawURL, "discovery/item") {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	q := u.Query()
	keys := []string{
		"source", "xhsshare", "app_platform", "ignoreEngage", "app_version",
		"share_from_user_hidden", "author_share", "shareRedId",
		"apptime", "share_id", "share_channel",
	}
	for _, k := range keys {
		q.Del(k)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (p *XiaohongshuParser) fetchHTML(pageURL string) (string, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return "", err
	}

	if strings.Contains(pageURL, "/explore/") || strings.Contains(pageURL, "xsec_source=pc") {
		req.Header.Set("User-Agent", xhsPCUA)
		req.Header.Set("Sec-CH-UA-Mobile", "?0")
		req.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	} else {
		req.Header.Set("User-Agent", xhsAndroidUA)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")

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

func (p *XiaohongshuParser) extractInitialState(html string) (map[string]interface{}, error) {
	markers := []string{
		"window.__SETUP_SERVER_STATE__",
		"window.__INITIAL_STATE__",
	}

	for _, marker := range markers {
		idx := strings.Index(html, marker)
		if idx == -1 {
			continue
		}

		jsonStart := strings.Index(html[idx:], "{")
		if jsonStart == -1 {
			continue
		}
		jsonStart += idx

		scriptEnd := strings.Index(html[idx:], "</script>")
		if scriptEnd == -1 {
			scriptEnd = len(html) - idx
		}

		braceCount := 0
		inString := false
		escapeNext := false
		jsonEnd := -1

		for i := jsonStart; i < idx+scriptEnd && i < len(html); i++ {
			c := html[i]
			if escapeNext {
				escapeNext = false
				continue
			}
			if c == '\\' {
				escapeNext = true
				continue
			}
			if c == '"' && !escapeNext {
				inString = !inString
				continue
			}
			if !inString {
				if c == '{' {
					braceCount++
				} else if c == '}' {
					braceCount--
					if braceCount == 0 {
						jsonEnd = i + 1
						break
					}
				}
			}
		}

		if jsonEnd == -1 {
			continue
		}

		jsonStr := html[jsonStart:jsonEnd]
		jsonStr = strings.ReplaceAll(jsonStr, ":undefined", ":null")

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			continue
		}
		return data, nil
	}

	return nil, errors.New("未找到 JSON 数据")
}

func (p *XiaohongshuParser) findNote(data map[string]interface{}) (map[string]interface{}, error) {
	// 移动端 __SETUP_SERVER_STATE__: data.LAUNCHER_SSR_STORE_PAGE_DATA.noteData
	if ssrObj, ok := data["LAUNCHER_SSR_STORE_PAGE_DATA"].(map[string]interface{}); ok {
		if note, ok := ssrObj["noteData"].(map[string]interface{}); ok && note != nil {
			if note["title"] != nil || note["desc"] != nil || note["video"] != nil || note["imageList"] != nil {
				return note, nil
			}
		}
	}

	// 移动端 __INITIAL_STATE__: data.noteData.data.noteData
	if noteDataObj, ok := data["noteData"].(map[string]interface{}); ok {
		if dataObj, ok := noteDataObj["data"].(map[string]interface{}); ok {
			if note, ok := dataObj["noteData"].(map[string]interface{}); ok && note != nil {
				if note["title"] != nil || note["desc"] != nil || note["video"] != nil || note["imageList"] != nil {
					return note, nil
				}
			}
		}
	}

	// PC端: data.note.noteDetailMap[noteId].note
	if noteObj, ok := data["note"].(map[string]interface{}); ok {
		if detailMap, ok := noteObj["noteDetailMap"].(map[string]interface{}); ok {
			for _, v := range detailMap {
				if detail, ok := v.(map[string]interface{}); ok {
					if note, ok := detail["note"].(map[string]interface{}); ok && note != nil {
						if note["title"] != nil || note["desc"] != nil || note["video"] != nil || note["imageList"] != nil {
							return note, nil
						}
					}
				}
			}
		}
	}

	return nil, errors.New("未找到笔记数据（小红书Web端可能需要登录或APP查看）")
}

func (p *XiaohongshuParser) parseVideo(note map[string]interface{}) (*VideoResult, string, error) {
	result := &VideoResult{Type: "video"}

	if title, ok := note["title"].(string); ok && title != "" {
		result.Title = strings.TrimSpace(title)
	}
	if desc, ok := note["desc"].(string); ok && desc != "" {
		desc = cleanXhsTopicTags(desc)
		if result.Title == "" {
			result.Title = desc
		} else {
			result.Title = result.Title + " " + desc
		}
	}

	if user, ok := note["user"].(map[string]interface{}); ok {
		if nick, ok := user["nickName"].(string); ok {
			result.Author = nick
		}
		if uid, ok := user["userId"].(string); ok {
			result.AuthorID = uid
		}
	}

	if interact, ok := note["interactInfo"].(map[string]interface{}); ok {
		if v, ok := interact["likedCount"].(string); ok {
			result.LikeCount = parseXhsCount(v)
		} else if v, ok := interact["likedCount"].(float64); ok {
			result.LikeCount = int64(v)
		}
		if v, ok := interact["commentCount"].(string); ok {
			result.CommentCount = parseXhsCount(v)
		} else if v, ok := interact["commentCount"].(float64); ok {
			result.CommentCount = int64(v)
		}
		if v, ok := interact["shareCount"].(string); ok {
			result.ViewCount = parseXhsCount(v)
		} else if v, ok := interact["shareCount"].(float64); ok {
			result.ViewCount = int64(v)
		}
	}

	if stats, ok := note["stats"].(map[string]interface{}); ok {
		if v, ok := stats["playCount"].(float64); ok && v > 0 {
			result.ViewCount = int64(v)
		} else if v, ok := stats["viewCount"].(float64); ok && v > 0 {
			result.ViewCount = int64(v)
		}
	}

	video, ok := note["video"].(map[string]interface{})
	if !ok {
		return nil, "", errors.New("未找到视频数据")
	}

	// 视频URL: video.url
	if vurl, ok := video["url"].(string); ok && vurl != "" {
		result.VideoURL = ensureHTTPS(vurl)
	}

	// 备选: video.downloadUrl
	if result.VideoURL == "" {
		if vurl, ok := video["downloadUrl"].(string); ok && vurl != "" {
			result.VideoURL = ensureHTTPS(vurl)
		}
	}

	// 备选: video.media
	if result.VideoURL == "" {
		if media, ok := video["media"].(map[string]interface{}); ok {
			if stream, ok := media["stream"].(map[string]interface{}); ok {
				for _, key := range []string{"h265", "h264", "av1"} {
					if arr, ok := stream[key].([]interface{}); ok && len(arr) > 0 {
						for _, item := range arr {
							s, ok := item.(map[string]interface{})
							if !ok {
								continue
							}
							if m, ok := s["masterUrl"].(string); ok && m != "" {
								result.VideoURL = ensureHTTPS(m)
								break
							}
							if m, ok := s["backupUrls"].([]interface{}); ok && len(m) > 0 {
								if backupURL, ok := m[0].(string); ok && backupURL != "" {
									result.VideoURL = ensureHTTPS(backupURL)
									break
								}
							}
						}
					}
					if result.VideoURL != "" {
						break
					}
				}
			}
		}
	}

	if result.VideoURL == "" {
		return nil, "", errors.New("未找到视频URL")
	}

	if dur, ok := video["duration"].(float64); ok {
		result.Duration = int(dur / 1000)
	}

	if imageList, ok := note["imageList"].([]interface{}); ok && len(imageList) > 0 {
		if firstImg, ok := imageList[0].(map[string]interface{}); ok {
			if u, ok := firstImg["url"].(string); ok && u != "" {
				result.CoverURL = xhsCoverURL(u)
			}
		}
	}

	return result, "video", nil
}

func (p *XiaohongshuParser) parseAtlas(note map[string]interface{}) (*AtlasResult, string, error) {
	result := &AtlasResult{Type: "atlas"}

	if title, ok := note["title"].(string); ok && title != "" {
		result.Title = strings.TrimSpace(title)
	}
	if desc, ok := note["desc"].(string); ok && desc != "" {
		desc = cleanXhsTopicTags(desc)
		if result.Title == "" {
			result.Title = desc
		} else {
			result.Title = result.Title + " " + desc
		}
	}

	if user, ok := note["user"].(map[string]interface{}); ok {
		if nick, ok := user["nickName"].(string); ok {
			result.Author = nick
		}
		if uid, ok := user["userId"].(string); ok {
			result.AuthorID = uid
		}
	}

	if interact, ok := note["interactInfo"].(map[string]interface{}); ok {
		if v, ok := interact["likedCount"].(string); ok {
			result.LikeCount = parseXhsCount(v)
		} else if v, ok := interact["likedCount"].(float64); ok {
			result.LikeCount = int64(v)
		}
		if v, ok := interact["commentCount"].(string); ok {
			result.CommentCount = parseXhsCount(v)
		} else if v, ok := interact["commentCount"].(float64); ok {
			result.CommentCount = int64(v)
		}
	}

	imageList, ok := note["imageList"].([]interface{})
	if !ok || len(imageList) == 0 {
		return nil, "", errors.New("未找到图片列表")
	}

	for _, item := range imageList {
		imgMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var imgInfo Image
		if u, ok := imgMap["url"].(string); ok && u != "" {
			imgInfo.URL = xhsCleanImageURL(u)
		}
		if imgInfo.URL != "" {
			if w, ok := imgMap["width"].(float64); ok {
				imgInfo.Width = int(w)
			}
			if h, ok := imgMap["height"].(float64); ok {
				imgInfo.Height = int(h)
			}
			result.Images = append(result.Images, imgInfo)
		}
	}

	if len(result.Images) == 0 {
		return nil, "", errors.New("未提取到有效图片")
	}

	return result, "atlas", nil
}

func cleanXhsTopicTags(text string) string {
	if text == "" {
		return text
	}
	re := regexp.MustCompile(`#([^#\[]+)\[话题\]#`)
	return re.ReplaceAllString(text, "#$1")
}

func parseXhsCount(s string) int64 {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(s)
	multiplier := 1.0
	switch {
	case strings.HasSuffix(s, "万"):
		multiplier = 10000
		s = strings.TrimSuffix(s, "万")
	case strings.HasSuffix(s, "亿"):
		multiplier = 100000000
		s = strings.TrimSuffix(s, "亿")
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return int64(v * multiplier)
}

func ensureHTTPS(u string) string {
	if strings.HasPrefix(u, "http://") {
		return "https://" + strings.TrimPrefix(u, "http://")
	}
	if strings.HasPrefix(u, "//") {
		return "https:" + u
	}
	return u
}

func xhsCleanImageURL(u string) string {
	if u == "" {
		return u
	}
	re := regexp.MustCompile(`/([^/!]+)(?:![^?]*)?$`)
	m := re.FindStringSubmatch(u)
	if len(m) < 2 {
		return u
	}
	fileId := m[1]
	if strings.Contains(u, "notes_pre_post/") {
		return "https://sns-img-qc.xhscdn.com/notes_pre_post/" + fileId
	}
	return "https://sns-img-qc.xhscdn.com/" + fileId
}

func xhsCoverURL(u string) string {
	return xhsCleanImageURL(u)
}
