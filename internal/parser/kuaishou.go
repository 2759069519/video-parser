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

// KuaishouParser 快手解析器
type KuaishouParser struct {
	client *http.Client
}

// NewKuaishouParser 创建快手解析器
func NewKuaishouParser() *KuaishouParser {
	return &KuaishouParser{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseURL 解析URL
func (p *KuaishouParser) ParseURL(url string) (interface{}, string, error) {
	// 获取重定向URL
	redirectURL, err := p.getRedirectURL(url)
	if err != nil {
		return nil, "", err
	}

	fmt.Printf("重定向URL: %s\n", redirectURL)

	// 转换URL为可解析的格式
	parseURL := redirectURL
	if strings.Contains(redirectURL, "chenzhongtech.com") {
		// 只转换非m开头的子域名，m开头的子域名证书兼容
		if !strings.Contains(redirectURL, ".m.chenzhongtech.com") && !strings.HasPrefix(redirectURL, "http://m.chenzhongtech.com") && !strings.HasPrefix(redirectURL, "https://m.chenzhongtech.com") {
			// gifshow域名转换
			if strings.Contains(redirectURL, "gifshow.com") {
				parseURL = strings.Replace(redirectURL, "m.gifshow.com", "m.chenzhongtech.com", 1)
				parseURL = strings.Replace(parseURL, "www.gifshow.com", "m.chenzhongtech.com", 1)
			}
		}
	} else if strings.Contains(redirectURL, "kuaishou.com") || strings.Contains(redirectURL, "kspkg.com") {
		// 将 www.kuaishou.com/short-video/ 替换为 m.chenzhongtech.com/fw/photo/
		parseURL = strings.Replace(redirectURL, "www.kuaishou.com/short-video/", "m.chenzhongtech.com/fw/photo/", 1)
		parseURL = strings.Replace(parseURL, "www.kuaishou.com/profile/", "m.chenzhongtech.com/fw/user/", 1)
	}

	// 移除查询参数
	if idx := strings.Index(parseURL, "?"); idx != -1 {
		parseURL = parseURL[:idx]
	}

	fmt.Printf("解析URL: %s\n", parseURL)

	// 判断类型
	if strings.Contains(redirectURL, "/fw/photo/") || strings.Contains(redirectURL, "/fw/long-video/") || strings.Contains(redirectURL, "/short-video/") {
		// 视频或图文 - 使用简化的 URL
		data, err := p.parsePage(parseURL)
		if err != nil {
			return nil, "", err
		}

		// 判断是视频还是图文
		if p.isVideo(data) {
			result := p.extractVideoInfo(data)
			return result, "video", nil
		} else if p.isAtlas(data) {
			result := p.extractAtlasInfo(data)
			return result, "atlas", nil
		}
		return nil, "", errors.New("无法识别内容类型")
	} else if strings.Contains(redirectURL, "/profile/") || strings.Contains(redirectURL, "/user/profile/") || strings.Contains(redirectURL, "/fw/user/") {
		// 主页 - 保留查询参数
		profileURL := redirectURL
		if strings.Contains(redirectURL, "kpfshanghai.m.chenzhongtech.com") {
			profileURL = strings.Replace(redirectURL, "kpfshanghai.m.chenzhongtech.com", "v.m.chenzhongtech.com", 1)
		}

		data, err := p.parsePage(profileURL)
		if err != nil {
			return nil, "", err
		}
		result := p.extractProfileInfo(data)
		return result, "profile", nil
	}

	return nil, "", fmt.Errorf("不支持的URL格式: %s", redirectURL)
}

// getRedirectURL 获取重定向URL
func (p *KuaishouParser) getRedirectURL(shortURL string) (string, error) {
	// 创建不跟随重定向的客户端
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

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查是否有重定向
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location != "" {
			return location, nil
		}
	}

	// 如果没有重定向，返回原始URL
	return shortURL, nil
}

// parsePage 解析页面
func (p *KuaishouParser) parsePage(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 使用与原项目相同的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://v.kuaishou.com/")

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

	// 查找 INIT_STATE
	start := strings.Index(html, "window.INIT_STATE = ")
	if start == -1 {
		return nil, errors.New("未找到INIT_STATE数据")
	}

	jsonStart := start + 20
	end := strings.Index(html[jsonStart:], "</script>")
	if end == -1 {
		return nil, errors.New("未找到script结束标签")
	}

	jsonStr := strings.TrimSpace(html[jsonStart : jsonStart+end])
	jsonStr = strings.TrimSuffix(jsonStr, ";")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return data, nil
}

// isVideo 判断是否为视频
func (p *KuaishouParser) isVideo(data map[string]interface{}) bool {
	for _, v := range data {
		if item, ok := v.(map[string]interface{}); ok {
			if photo, ok := item["photo"].(map[string]interface{}); ok {
				if photoType, ok := photo["photoType"].(string); ok {
					return photoType == "VIDEO"
				}
			}
		}
	}
	return false
}

// isAtlas 判断是否为图文
func (p *KuaishouParser) isAtlas(data map[string]interface{}) bool {
	for _, v := range data {
		if item, ok := v.(map[string]interface{}); ok {
			if photo, ok := item["photo"].(map[string]interface{}); ok {
				if photoType, ok := photo["photoType"].(string); ok {
					return photoType == "HORIZONTAL_ATLAS" || photoType == "VERTICAL_ATLAS"
				}
			}
		}
	}
	return false
}

// extractVideoInfo 提取视频信息
func (p *KuaishouParser) extractVideoInfo(data map[string]interface{}) *VideoResult {
	result := &VideoResult{
		Type: "video",
	}

	for _, v := range data {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		photo, ok := item["photo"].(map[string]interface{})
		if !ok {
			continue
		}

		if photoType, ok := photo["photoType"].(string); ok && photoType == "VIDEO" {
			if caption, ok := photo["caption"].(string); ok {
				result.Title = caption
			}
			if userName, ok := photo["userName"].(string); ok {
				result.Author = userName
			}
			if userID, ok := photo["userId"].(string); ok {
				result.AuthorID = userID
			}
			if likeCount, ok := photo["likeCount"].(float64); ok {
				result.LikeCount = int64(likeCount)
			}
			if commentCount, ok := photo["commentCount"].(float64); ok {
				result.CommentCount = int64(commentCount)
			}
			if viewCount, ok := photo["viewCount"].(float64); ok {
				result.ViewCount = int64(viewCount)
			}
			if duration, ok := photo["duration"].(float64); ok {
				result.Duration = int(duration)
			}

			// 获取视频URL
			if manifest, ok := photo["manifest"].(map[string]interface{}); ok {
				if adaptationSet, ok := manifest["adaptationSet"].([]interface{}); ok && len(adaptationSet) > 0 {
					if adapt, ok := adaptationSet[0].(map[string]interface{}); ok {
						if representation, ok := adapt["representation"].([]interface{}); ok && len(representation) > 0 {
							// 找最高码率
							var bestURL string
							var bestBitrate float64
							for _, rep := range representation {
								if repMap, ok := rep.(map[string]interface{}); ok {
									if url, ok := repMap["url"].(string); ok {
										if bitrate, ok := repMap["avgBitrate"].(float64); ok {
											if bitrate > bestBitrate {
												bestBitrate = bitrate
												bestURL = url
											}
										}
									}
								}
							}
							if bestURL != "" {
								result.VideoURL = bestURL
							}
						}
					}
				}
			}

			// 获取封面URL
			if coverUrls, ok := photo["coverUrls"].([]interface{}); ok && len(coverUrls) > 0 {
				if cover, ok := coverUrls[0].(map[string]interface{}); ok {
					if url, ok := cover["url"].(string); ok {
						result.CoverURL = url
					}
				}
			}

			break
		}
	}

	return result
}

// extractAtlasInfo 提取图文信息
func (p *KuaishouParser) extractAtlasInfo(data map[string]interface{}) *AtlasResult {
	result := &AtlasResult{
		Type: "atlas",
	}

	for _, v := range data {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		photo, ok := item["photo"].(map[string]interface{})
		if !ok {
			continue
		}

		if photoType, ok := photo["photoType"].(string); ok && (photoType == "HORIZONTAL_ATLAS" || photoType == "VERTICAL_ATLAS") {
			if caption, ok := photo["caption"].(string); ok {
				result.Title = caption
			}
			if userName, ok := photo["userName"].(string); ok {
				result.Author = userName
			}
			if userID, ok := photo["userId"].(string); ok {
				result.AuthorID = userID
			}
			if likeCount, ok := photo["likeCount"].(float64); ok {
				result.LikeCount = int64(likeCount)
			}
			if commentCount, ok := photo["commentCount"].(float64); ok {
				result.CommentCount = int64(commentCount)
			}

			// 获取图片列表
			if extParams, ok := photo["ext_params"].(map[string]interface{}); ok {
				if atlas, ok := extParams["atlas"].(map[string]interface{}); ok {
					if list, ok := atlas["list"].([]interface{}); ok {
						cdn := "p3.a.yximgs.com"
						if cdnList, ok := atlas["cdn"].([]interface{}); ok && len(cdnList) > 0 {
							if c, ok := cdnList[0].(string); ok {
								cdn = c
							}
						}

						sizes := make([]map[string]interface{}, 0)
						if sizeList, ok := atlas["size"].([]interface{}); ok {
							for _, s := range sizeList {
								if sizeMap, ok := s.(map[string]interface{}); ok {
									sizes = append(sizes, sizeMap)
								}
							}
						}

						for i, path := range list {
							if pathStr, ok := path.(string); ok {
								img := Image{
									URL: fmt.Sprintf("https://%s%s", cdn, pathStr),
								}
								if i < len(sizes) {
									if w, ok := sizes[i]["w"].(float64); ok {
										img.Width = int(w)
									}
									if h, ok := sizes[i]["h"].(float64); ok {
										img.Height = int(h)
									}
								}
								result.Images = append(result.Images, img)
							}
						}
					}
				}
			}

			// 如果没有获取到图片，尝试从coverUrls获取
			if len(result.Images) == 0 {
				if coverUrls, ok := photo["coverUrls"].([]interface{}); ok {
					for _, cover := range coverUrls {
						if coverMap, ok := cover.(map[string]interface{}); ok {
							if url, ok := coverMap["url"].(string); ok {
								result.Images = append(result.Images, Image{URL: url})
							}
						}
					}
				}
			}

			break
		}
	}

	return result
}

// extractProfileInfo 提取主页信息
func (p *KuaishouParser) extractProfileInfo(data map[string]interface{}) *ProfileResult {
	result := &ProfileResult{
		Type: "profile",
	}

	// 先查找作品作者ID
	feedAuthorID := p.findFeedsAuthor(data)

	// 查找用户信息
	for _, v := range data {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		if userProfile, ok := item["userProfile"].(map[string]interface{}); ok {
			if profile, ok := userProfile["profile"].(map[string]interface{}); ok {
				// 支持字符串和数字类型的 user_id
				userID := getString(profile, "user_id")
				if userID == "" {
					if uid, ok := profile["user_id"].(float64); ok && uid > 0 {
						userID = fmt.Sprintf("%.0f", uid)
					}
				}

				// 如果找到了作品作者ID，只匹配作者信息
				if feedAuthorID != "" && userID != feedAuthorID {
					continue
				}

				result.UserID = userID
				result.UserName = getString(profile, "user_name")
				result.Avatar = getString(profile, "headurl")
				result.Description = getString(profile, "user_text")

				if ownerCount, ok := userProfile["ownerCount"].(map[string]interface{}); ok {
					if fan, ok := ownerCount["fan"].(float64); ok {
						result.FanCount = int64(fan)
					}
					if follow, ok := ownerCount["follow"].(float64); ok {
						result.FollowCount = int64(follow)
					}
					if photo, ok := ownerCount["photo"].(float64); ok {
						result.PhotoCount = int64(photo)
					}
				}
				break
			}
		}
	}

	// 查找作品列表
	p.findFeeds(data, result)

	return result
}

// findFeedsAuthor 查找作品作者ID
func (p *KuaishouParser) findFeedsAuthor(data map[string]interface{}) string {
	var authorID string
	p.findFeedsAuthorRecursive(data, &authorID, 0)
	return authorID
}

// findFeedsAuthorRecursive 递归查找作品作者ID
func (p *KuaishouParser) findFeedsAuthorRecursive(obj interface{}, authorID *string, depth int) {
	if depth > 5 || *authorID != "" {
		return
	}

	// 如果是数组，检查是否有 userId
	if arr, ok := obj.([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			if _, hasCoverUrls := first["coverUrls"]; hasCoverUrls {
				// 支持字符串和数字类型的 userId
				if uid, ok := first["userId"].(string); ok && uid != "" {
					*authorID = uid
					return
				}
				if uid, ok := first["userId"].(float64); ok && uid > 0 {
					*authorID = fmt.Sprintf("%.0f", uid)
					return
				}
			}
		}
	}

	// 如果是对象，遍历键
	if m, ok := obj.(map[string]interface{}); ok {
		for key, v := range m {
			if key == "feeds" {
				if feeds, ok := v.([]interface{}); ok && len(feeds) > 0 {
					if first, ok := feeds[0].(map[string]interface{}); ok {
						// 支持字符串和数字类型的 userId
						if uid, ok := first["userId"].(string); ok && uid != "" {
							*authorID = uid
							return
						}
						if uid, ok := first["userId"].(float64); ok && uid > 0 {
							*authorID = fmt.Sprintf("%.0f", uid)
							return
						}
					}
				}
			}
			if v != nil {
				p.findFeedsAuthorRecursive(v, authorID, depth+1)
				if *authorID != "" {
					return
				}
			}
		}
	}
}

// findFeeds 查找作品列表
func (p *KuaishouParser) findFeeds(data map[string]interface{}, result *ProfileResult) {
	// 递归查找 feeds 字段
	p.findFeedsRecursive(data, result, 0)
}

// findFeedsRecursive 递归查找 feeds
func (p *KuaishouParser) findFeedsRecursive(obj interface{}, result *ProfileResult, depth int) {
	if depth > 5 || result.Photos != nil {
		return
	}

	// 如果是数组，检查是否是作品列表
	if arr, ok := obj.([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			if _, hasCoverUrls := first["coverUrls"]; hasCoverUrls {
				p.extractPhotos(arr, result)
				return
			}
		}
	}

	// 如果是对象，遍历键
	if m, ok := obj.(map[string]interface{}); ok {
		for key, v := range m {
			if key == "feeds" {
				if feeds, ok := v.([]interface{}); ok && len(feeds) > 0 {
					p.extractPhotos(feeds, result)
					if len(result.Photos) > 0 {
						return
					}
				}
			}
			if v != nil {
				p.findFeedsRecursive(v, result, depth+1)
				if len(result.Photos) > 0 {
					return
				}
			}
		}
	}
}

// extractPhotos 提取作品信息
func (p *KuaishouParser) extractPhotos(feeds []interface{}, result *ProfileResult) {
	for _, feed := range feeds {
		feedMap, ok := feed.(map[string]interface{})
		if !ok {
			continue
		}

		photo := Photo{
			PhotoID: getString(feedMap, "photoId"),
			Type:    getString(feedMap, "photoType"),
			Caption: getString(feedMap, "caption"),
		}

		if coverUrls, ok := feedMap["coverUrls"].([]interface{}); ok && len(coverUrls) > 0 {
			if cover, ok := coverUrls[0].(map[string]interface{}); ok {
				photo.CoverURL = getString(cover, "url")
			}
		}

		if likeCount, ok := feedMap["likeCount"].(float64); ok {
			photo.LikeCount = int64(likeCount)
		}
		if commentCount, ok := feedMap["commentCount"].(float64); ok {
			photo.CommentCount = int64(commentCount)
		}
		if viewCount, ok := feedMap["viewCount"].(float64); ok {
			photo.ViewCount = int64(viewCount)
		}
		if duration, ok := feedMap["duration"].(float64); ok {
			photo.Duration = int(duration)
		}

		result.Photos = append(result.Photos, photo)
	}
}

// getString 从map中获取字符串
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// FetchVideoURL 获取单个作品的视频URL
func (p *KuaishouParser) FetchVideoURL(photoID string) (string, error) {
	url := fmt.Sprintf("https://v.m.chenzhongtech.com/fw/photo/%s?shareId=0&userId=0", photoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// 使用与原项目相同的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://v.kuaishou.com/")

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

	html := string(body)

	// 查找INIT_STATE
	start := strings.Index(html, "window.INIT_STATE = ")
	if start == -1 {
		return "", errors.New("未找到INIT_STATE数据")
	}

	jsonStart := start + 20
	end := strings.Index(html[jsonStart:], "</script>")
	if end == -1 {
		return "", errors.New("未找到script结束标签")
	}

	jsonStr := strings.TrimSpace(html[jsonStart : jsonStart+end])
	jsonStr = strings.TrimSuffix(jsonStr, ";")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}

	// 查找视频URL
	for _, v := range data {
		if item, ok := v.(map[string]interface{}); ok {
			if photo, ok := item["photo"].(map[string]interface{}); ok {
				if manifest, ok := photo["manifest"].(map[string]interface{}); ok {
					if adaptationSet, ok := manifest["adaptationSet"].([]interface{}); ok && len(adaptationSet) > 0 {
						if adapt, ok := adaptationSet[0].(map[string]interface{}); ok {
							if representation, ok := adapt["representation"].([]interface{}); ok && len(representation) > 0 {
								if rep, ok := representation[0].(map[string]interface{}); ok {
									if url, ok := rep["url"].(string); ok {
										return url, nil
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return "", errors.New("未找到视频URL")
}

// FetchAtlasImages 获取图文作品的图片列表
func (p *KuaishouParser) FetchAtlasImages(photoID string) ([]Image, error) {
	url := fmt.Sprintf("https://v.m.chenzhongtech.com/fw/photo/%s?shareId=0&userId=0", photoID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 使用与原项目相同的请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Referer", "https://v.kuaishou.com/")

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

	// 查找INIT_STATE
	start := strings.Index(html, "window.INIT_STATE = ")
	if start == -1 {
		return nil, errors.New("未找到INIT_STATE数据")
	}

	jsonStart := start + 20
	end := strings.Index(html[jsonStart:], "</script>")
	if end == -1 {
		return nil, errors.New("未找到script结束标签")
	}

	jsonStr := strings.TrimSpace(html[jsonStart : jsonStart+end])
	jsonStr = strings.TrimSuffix(jsonStr, ";")

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	// 查找图片列表
	for _, v := range data {
		if item, ok := v.(map[string]interface{}); ok {
			if photo, ok := item["photo"].(map[string]interface{}); ok {
				// 检查是否是图文类型
				photoType := getString(photo, "photoType")
				if photoType != "HORIZONTAL_ATLAS" {
					return nil, errors.New("不是图文类型")
				}

				// 获取图片列表
				if extParams, ok := photo["ext_params"].(map[string]interface{}); ok {
					if atlas, ok := extParams["atlas"].(map[string]interface{}); ok {
						if list, ok := atlas["list"].([]interface{}); ok {
							cdn := "p3.a.yximgs.com"
							if cdnList, ok := atlas["cdn"].([]interface{}); ok && len(cdnList) > 0 {
								if c, ok := cdnList[0].(string); ok {
									cdn = c
								}
							}

							sizes := make([]map[string]interface{}, 0)
							if sizeList, ok := atlas["size"].([]interface{}); ok {
								for _, s := range sizeList {
									if sizeMap, ok := s.(map[string]interface{}); ok {
										sizes = append(sizes, sizeMap)
									}
								}
							}

							var images []Image
							for i, path := range list {
								if pathStr, ok := path.(string); ok {
									img := Image{
										URL: fmt.Sprintf("https://%s%s", cdn, pathStr),
									}
									if i < len(sizes) {
										if w, ok := sizes[i]["w"].(float64); ok {
											img.Width = int(w)
										}
										if h, ok := sizes[i]["h"].(float64); ok {
											img.Height = int(h)
										}
									}
									images = append(images, img)
								}
							}
							return images, nil
						}
					}
				}

				// 如果没有获取到图片，尝试从coverUrls获取
				if coverUrls, ok := photo["coverUrls"].([]interface{}); ok {
					var images []Image
					for _, cover := range coverUrls {
						if coverMap, ok := cover.(map[string]interface{}); ok {
							if url, ok := coverMap["url"].(string); ok {
								images = append(images, Image{URL: url})
							}
						}
					}
					if len(images) > 0 {
						return images, nil
					}
				}
			}
		}
	}

	return nil, errors.New("未找到图片列表")
}

// ExtractPhotoID 从URL中提取photoID
func ExtractPhotoID(url string) string {
	// 匹配 /fw/photo/xxx 或 /short-video/xxx
	re := regexp.MustCompile(`/(?:fw/photo|short-video)/(\w+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
