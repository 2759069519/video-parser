package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"video-parser/internal/service"
	"video-parser/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxBodySize       = 50 * 1024 * 1024
	maxDownloadSize   = 100 * 1024 * 1024
	maxImageSize      = 10 * 1024 * 1024
	allowedImageTypes = "image/jpeg,image/png,image/webp,image/avif"
)

var downloadClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		if _, err := utils.ResourcePlatform(req.URL.String()); err != nil {
			return err
		}
		return nil
	},
}

type VideoHandler struct {
	service *service.VideoService
}

func NewVideoHandler() *VideoHandler {
	return &VideoHandler{
		service: service.NewVideoService(),
	}
}

type ParseRequest struct {
	URL string `json:"url" binding:"required"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Type    string      `json:"type,omitempty"`
	Error   string      `json:"error,omitempty"`
}

var urlRe = regexp.MustCompile(`https?://[^\s]*(?:kuaishou\.com|gifshow\.com|chenzhongtech\.com|kspkg\.com|douyin\.com|iesdouyin\.com|xiaohongshu\.com|xhslink\.com|miyoushe\.com|doubao\.com)[^\s]*`)

func extractURL(input string) string {
	if urlRe.MatchString(input) {
		return urlRe.FindString(input)
	}
	return input
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "\n", "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "file"
	}
	return name
}

func isAllowedImageType(contentType string) bool {
	ct := strings.ToLower(contentType)
	for _, allowed := range strings.Split(allowedImageTypes, ",") {
		if strings.Contains(ct, strings.TrimSpace(allowed)) {
			return true
		}
	}
	return false
}

func (h *VideoHandler) Parse(c *gin.Context) {
	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "请求参数错误: " + err.Error()})
		return
	}
	req.URL = extractURL(req.URL)
	
	result, contentType, err := h.service.ParseURL(req.URL, c.ClientIP(), c.Request.UserAgent())
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "解析失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: result, Type: contentType})
}

func (h *VideoHandler) GetVideo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "无效的ID"})
		return
	}
	video, err := h.service.GetVideoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{Success: false, Error: "视频不存在"})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: video, Type: "video"})
}

func (h *VideoHandler) GetAtlas(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "无效的ID"})
		return
	}
	atlas, err := h.service.GetAtlasByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{Success: false, Error: "图文不存在"})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: atlas, Type: "atlas"})
}

func (h *VideoHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "无效的ID"})
		return
	}
	profile, err := h.service.GetProfileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{Success: false, Error: "主页不存在"})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: profile, Type: "profile"})
}

func (h *VideoHandler) GetRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	platform := c.Query("platform")
	status := c.Query("status")
	contentType := c.Query("type")

	result, err := h.service.GetParseRecords(page, pageSize, platform, status, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "获取记录失败"})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: result})
}

type FetchVideoURLRequest struct {
	PhotoID  string `json:"photo_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}

func (h *VideoHandler) FetchVideoURL(c *gin.Context) {
	var req FetchVideoURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "请求参数错误: " + err.Error()})
		return
	}
	videoURL, err := h.service.FetchVideoURL(req.PhotoID, req.Platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "获取视频URL失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: map[string]string{"video_url": videoURL}})
}

type FetchAtlasImagesRequest struct {
	PhotoID  string `json:"photo_id" binding:"required"`
	Platform string `json:"platform" binding:"required"`
}

func (h *VideoHandler) FetchAtlasImages(c *gin.Context) {
	var req FetchAtlasImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "请求参数错误: " + err.Error()})
		return
	}
	images, err := h.service.FetchAtlasImages(req.PhotoID, req.Platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "获取图片失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, ParseResponse{Success: true, Data: map[string]interface{}{"images": images}})
}

func (h *VideoHandler) ProxyImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "缺少url参数"})
		return
	}
	platform, err := utils.ResourcePlatform(imageURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "不支持的URL: " + err.Error()})
		return
	}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "创建请求失败"})
		return
	}

	switch platform {
	case "xhs":
		req.Header.Set("Referer", "https://www.xiaohongshu.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	case "doubao":
		req.Header.Set("Referer", "https://www.doubao.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	case "douyin":
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)")
	case "kuaishou":
		req.Header.Set("Referer", "https://www.kuaishou.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)")
	case "miyoushe":
		req.Header.Set("Referer", "https://m.miyoushe.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36")
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		log.Printf("[ProxyImage] download failed url=%s platform=%s err=%v", imageURL, platform, err)
		c.JSON(http.StatusBadGateway, ParseResponse{Success: false, Error: "下载图片失败"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ProxyImage] remote status abnormal url=%s platform=%s status=%d", imageURL, platform, resp.StatusCode)
		c.JSON(http.StatusBadGateway, ParseResponse{Success: false, Error: fmt.Sprintf("远程返回状态码: %d", resp.StatusCode)})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !isAllowedImageType(contentType) {
		log.Printf("[ProxyImage] disallowed content type url=%s platform=%s contentType=%s", imageURL, platform, contentType)
		c.JSON(http.StatusUnsupportedMediaType, ParseResponse{Success: false, Error: "不支持的图片类型: " + contentType})
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")

	limitedReader := io.LimitReader(resp.Body, maxImageSize)
	written, _ := io.Copy(c.Writer, limitedReader)
	if written >= maxImageSize {
		log.Printf("[ProxyImage] image too large url=%s platform=%s size=%d", imageURL, platform, written)
	}
}

func (h *VideoHandler) Download(c *gin.Context) {
	videoURL := c.Query("url")
	filename := c.Query("filename")
	if videoURL == "" {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "缺少url参数"})
		return
	}

	if err := utils.IsAllowedDownloadURL(videoURL); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "不允许下载此URL: " + err.Error()})
		return
	}

	platform, err := utils.ResourcePlatform(videoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{Success: false, Error: "不支持的资源URL: " + err.Error()})
		return
	}

	filename = sanitizeFilename(filename)
	if filename == "" {
		if platform == "xhs" {
			filename = "file"
		} else {
			filename = "video.mp4"
		}
	}

	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "创建请求失败: " + err.Error()})
		return
	}

	switch platform {
	case "xhs":
		req.Header.Set("Referer", "https://www.xiaohongshu.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	case "douyin":
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15")
	case "kuaishou":
		req.Header.Set("Referer", "https://www.kuaishou.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15")
	case "miyoushe":
		req.Header.Set("Referer", "https://m.miyoushe.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36")
	case "doubao":
		req.Header.Set("Referer", "https://www.doubao.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	default:
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		log.Printf("[Download] download failed url=%s platform=%s err=%v", videoURL, platform, err)
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: "下载失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[Download] remote status abnormal url=%s platform=%s status=%d", videoURL, platform, resp.StatusCode)
		c.JSON(http.StatusInternalServerError, ParseResponse{Success: false, Error: fmt.Sprintf("远程返回状态码: %d", resp.StatusCode)})
		return
	}

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil && size > maxDownloadSize {
			log.Printf("[Download] file too large url=%s platform=%s size=%d", videoURL, platform, size)
			c.JSON(http.StatusRequestEntityTooLarge, ParseResponse{Success: false, Error: "文件过大"})
			return
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Status(http.StatusOK)

	limitedReader := io.LimitReader(resp.Body, maxDownloadSize)
	written, _ := io.Copy(c.Writer, limitedReader)
	if written >= maxDownloadSize {
		log.Printf("[Download] download exceeded limit url=%s platform=%s written=%d", videoURL, platform, written)
	}
}
