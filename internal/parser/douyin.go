package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type DouyinParser struct {
	client *http.Client
}

func NewDouyinParser() *DouyinParser {
	return &DouyinParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *DouyinParser) ParseURL(rawURL string) (interface{}, string, error) {
	redirectURL, err := p.getRedirectURL(rawURL)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("抖音重定向URL: %s\n", redirectURL)

	if strings.Contains(redirectURL, "/note/") || strings.Contains(redirectURL, "/slides/") {
		return p.parseAtlas(redirectURL)
	}

	return p.parseVideo(redirectURL)
}

func (p *DouyinParser) getRedirectURL(shortURL string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", shortURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.douyin.com/")

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

	return shortURL, nil
}

func (p *DouyinParser) parseVideo(redirectURL string) (*VideoResult, string, error) {
	videoID := p.extractVideoID(redirectURL)
	if videoID == "" {
		return nil, "", errors.New("无法提取视频ID")
	}

	var pageURL string
	if strings.Contains(redirectURL, "v.douyin.com") || strings.Contains(redirectURL, "www.iesdouyin.com") {
		pageURL = fmt.Sprintf("https://www.iesdouyin.com/share/video/%s/", videoID)
	} else {
		pageURL = redirectURL
	}
	_ = pageURL

	data, err := p.fetchPageData(pageURL)
	if err != nil {
		return nil, "", err
	}

	result := &VideoResult{Type: "video"}

	if desc, ok := data["desc"].(string); ok {
		result.Title = desc
	}

	if author, ok := data["author"].(map[string]interface{}); ok {
		if nickname, ok := author["nickname"].(string); ok {
			result.Author = nickname
		}
		if uniqueID, ok := author["unique_id"].(string); ok {
			result.AuthorID = uniqueID
		} else if awemeID, ok := author["aweme_id"].(string); ok {
			result.AuthorID = awemeID
		}
	}

	if video, ok := data["video"].(map[string]interface{}); ok {
		// 尝试无水印的 play_addr
		if playAddr, ok := video["play_addr"].(map[string]interface{}); ok {
			if urlList, ok := playAddr["url_list"].([]interface{}); ok && len(urlList) > 0 {
				for _, u := range urlList {
					if urlStr, ok := u.(string); ok && urlStr != "" && !strings.Contains(urlStr, "playwm") {
						result.VideoURL = urlStr
						break
					}
				}
			}
		}
		// 尝试 bit_rate 数组中的高质量视频URL
		if result.VideoURL == "" {
			if bitRate, ok := video["bit_rate"].([]interface{}); ok && len(bitRate) > 0 {
				for i := len(bitRate) - 1; i >= 0; i-- {
					br, _ := bitRate[i].(map[string]interface{})
					if br == nil {
						continue
					}
					if playAddr, ok := br["play_addr"].(map[string]interface{}); ok {
						if urlList, ok := playAddr["url_list"].([]interface{}); ok && len(urlList) > 0 {
							for _, u := range urlList {
								if urlStr, ok := u.(string); ok && urlStr != "" && !strings.Contains(urlStr, "playwm") {
									result.VideoURL = urlStr
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
		// 尝试 play_addr_lowbr (低码率无水印)
		if result.VideoURL == "" {
			if lowBr, ok := video["play_addr_lowbr"].(map[string]interface{}); ok {
				if urlList, ok := lowBr["url_list"].([]interface{}); ok && len(urlList) > 0 {
					for _, u := range urlList {
						if urlStr, ok := u.(string); ok && urlStr != "" && !strings.Contains(urlStr, "playwm") {
							result.VideoURL = urlStr
							break
						}
					}
				}
			}
		}
		// 如果都没找到，回退到 uri 拼接 (使用 snssdk 官方接口，无水印)
		if result.VideoURL == "" {
			if playAddr, ok := video["play_addr"].(map[string]interface{}); ok {
				if uri, ok := playAddr["uri"].(string); ok {
					if strings.HasPrefix(uri, "http") {
						result.VideoURL = uri
					} else {
						result.VideoURL = fmt.Sprintf("https://aweme.snssdk.com/aweme/v1/play/?video_id=%s", uri)
					}
				}
			}
		}
		if cover, ok := video["cover"].(map[string]interface{}); ok {
			if urlList, ok := cover["url_list"].([]interface{}); ok && len(urlList) > 0 {
				if url, ok := urlList[0].(string); ok {
					result.CoverURL = url
				}
			}
		}
		if duration, ok := video["duration"].(float64); ok {
			result.Duration = int(duration / 1000)
		}
	}

	if statistics, ok := data["statistics"].(map[string]interface{}); ok {
		if likeCount, ok := statistics["digg_count"].(float64); ok {
			result.LikeCount = int64(likeCount)
		}
		if commentCount, ok := statistics["comment_count"].(float64); ok {
			result.CommentCount = int64(commentCount)
		}
		if shareCount, ok := statistics["share_count"].(float64); ok {
			result.ViewCount = int64(shareCount)
		}
	}

	if result.VideoURL == "" {
		return nil, "", errors.New("未找到视频地址")
	}

	// 如果是抖音play接口URL，服务端带Referer follow重定向得到真实CDN URL
	if strings.Contains(result.VideoURL, "aweme.snssdk.com") || strings.Contains(result.VideoURL, "aweme.qichuangtianxia.com") {
		if realURL, err := p.FetchRealVideoURL(result.VideoURL); err == nil && realURL != "" {
			result.VideoURL = realURL
		}
	}

	return result, "video", nil
}

func (p *DouyinParser) parseAtlas(redirectURL string) (*AtlasResult, string, error) {
	atlasID := p.extractAtlasID(redirectURL)
	if atlasID == "" {
		return nil, "", errors.New("无法提取图文ID")
	}

	// slides 路径用 /note/ 拼接 (slides 的 HTML 没有 loaderData，但 /note/ 路径有)
	var pageURL string
	if strings.Contains(redirectURL, "v.douyin.com") || strings.Contains(redirectURL, "www.iesdouyin.com") {
		pageURL = fmt.Sprintf("https://www.iesdouyin.com/share/note/%s/", atlasID)
	} else {
		pageURL = redirectURL
	}

	data, err := p.fetchPageData(pageURL)
	if err != nil {
		return nil, "", err
	}

	result := &AtlasResult{Type: "atlas"}

	if desc, ok := data["desc"].(string); ok {
		result.Title = desc
	}

	if author, ok := data["author"].(map[string]interface{}); ok {
		if nickname, ok := author["nickname"].(string); ok {
			result.Author = nickname
		}
		if uniqueID, ok := author["unique_id"].(string); ok {
			result.AuthorID = uniqueID
		} else if awemeID, ok := author["aweme_id"].(string); ok {
			result.AuthorID = awemeID
		}
	}

	if images, ok := data["images"].([]interface{}); ok {
		for _, img := range images {
			if imgMap, ok := img.(map[string]interface{}); ok {
				var imgInfo Image
				if urlList, ok := imgMap["url_list"].([]interface{}); ok {
					for _, u := range urlList {
						if urlStr, ok := u.(string); ok {
							imgInfo.URL = urlStr
							break
						}
					}
				}
				if width, ok := imgMap["width"].(float64); ok {
					imgInfo.Width = int(width)
				}
				if height, ok := imgMap["height"].(float64); ok {
					imgInfo.Height = int(height)
				}
				if imgInfo.URL != "" {
					result.Images = append(result.Images, imgInfo)
				}
			}
		}
	}

	if statistics, ok := data["statistics"].(map[string]interface{}); ok {
		if likeCount, ok := statistics["digg_count"].(float64); ok {
			result.LikeCount = int64(likeCount)
		}
		if commentCount, ok := statistics["comment_count"].(float64); ok {
			result.CommentCount = int64(commentCount)
		}
	}

	if len(result.Images) == 0 {
		return nil, "", errors.New("未找到图片")
	}

	return result, "atlas", nil
}

func (p *DouyinParser) fetchPageData(pageURL string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://www.douyin.com/")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, err
	}

	html := string(body)

	start := strings.Index(html, "window._ROUTER_DATA")
	if start == -1 {
		return nil, errors.New("未找到页面数据")
	}

	eqIdx := strings.Index(html[start:], "=")
	if eqIdx == -1 {
		return nil, errors.New("未找到等号")
	}
	start = start + eqIdx + 1
	for start < len(html) && (html[start] == ' ' || html[start] == '\n' || html[start] == '\t') {
		start++
	}

	end := strings.Index(html[start:], "</script>")
	if end == -1 {
		return nil, errors.New("未找到数据结束位置")
	}

	jsonStr := strings.TrimSpace(html[start : start+end])
	jsonStr = strings.TrimSuffix(jsonStr, ";")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		legacyJSON := strings.ReplaceAll(jsonStr, "\\\"", "\"")
		legacyJSON = strings.ReplaceAll(legacyJSON, "\\n", "")
		legacyJSON = strings.ReplaceAll(legacyJSON, "\\/", "/")
		legacyJSON = strings.ReplaceAll(legacyJSON, "\\u002F", "/")
		if legacyErr := json.Unmarshal([]byte(legacyJSON), &data); legacyErr != nil {
			return nil, fmt.Errorf("解析JSON失败: %v", err)
		}
	}

	if loaderData, ok := data["loaderData"].(map[string]interface{}); ok {
		// 视频: video_(id)/page -> videoInfoRes
		if videoPage, ok := loaderData["video_(id)/page"].(map[string]interface{}); ok {
			if videoInfoRes, ok := videoPage["videoInfoRes"].(map[string]interface{}); ok {
				if itemList, ok := videoInfoRes["item_list"].([]interface{}); ok && len(itemList) > 0 {
					if item, ok := itemList[0].(map[string]interface{}); ok {
						return item, nil
					}
				}
			}
		}

		// 图文: note_(id)/page -> videoInfoRes
		if notePage, ok := loaderData["note_(id)/page"].(map[string]interface{}); ok {
			if videoInfoRes, ok := notePage["videoInfoRes"].(map[string]interface{}); ok {
				if itemList, ok := videoInfoRes["item_list"].([]interface{}); ok && len(itemList) > 0 {
					if item, ok := itemList[0].(map[string]interface{}); ok {
						return item, nil
					}
				}
			}
		}
	}

	return nil, errors.New("未找到页面数据")
}

func (p *DouyinParser) extractVideoID(url string) string {
	re := regexp.MustCompile("/video/(\\d+)")
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (p *DouyinParser) extractAtlasID(url string) string {
	re := regexp.MustCompile("/(?:note|slides)/(\\d+)")
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// FetchRealVideoURL 跟随抖音play接口的302重定向，返回真实CDN URL
func (p *DouyinParser) FetchRealVideoURL(playURL string) (string, error) {
	if !strings.Contains(playURL, "aweme.snssdk.com") && !strings.Contains(playURL, "aweme.qichuangtianxia.com") {
		return playURL, nil
	}

	// 如果是 playwm URL（带水印），替换为 play URL（无水印）
	if strings.Contains(playURL, "/playwm/") {
		playURL = strings.Replace(playURL, "/playwm/", "/play/", 1)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest("GET", playURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Referer", "https://www.douyin.com/")

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

	return playURL, nil
}
