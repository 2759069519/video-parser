package handler

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"video-parser/internal/service"
	"video-parser/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxBodySize = 50 * 1024 * 1024
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

var urlRe = regexp.MustCompile(`https?://[^\s]*(?:kuaishou\.com|gifshow\.com|chenzhongtech\.com|kspkg\.com|douyin\.com|iesdouyin\.com|xiaohongshu\.com|xhslink\.com|miyoushe\.com)[^\s]*`)

func extractURL(input string) string {
	if urlRe.MatchString(input) {
		return urlRe.FindString(input)
	}
	return input
}

func (h *VideoHandler) Parse(c *gin.Context) {
	var req ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	req.URL = extractURL(req.URL)

	result, contentType, err := h.service.ParseURL(req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "解析失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data:    result,
		Type:    contentType,
	})
}

func (h *VideoHandler) GetVideo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "无效的ID",
		})
		return
	}

	video, err := h.service.GetVideoByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{
			Success: false,
			Error:   "视频不存在",
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data:    video,
		Type:    "video",
	})
}

func (h *VideoHandler) GetAtlas(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "无效的ID",
		})
		return
	}

	atlas, err := h.service.GetAtlasByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{
			Success: false,
			Error:   "图文不存在",
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data:    atlas,
		Type:    "atlas",
	})
}

func (h *VideoHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "无效的ID",
		})
		return
	}

	profile, err := h.service.GetProfileByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, ParseResponse{
			Success: false,
			Error:   "主页不存在",
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data:    profile,
		Type:    "profile",
	})
}

func (h *VideoHandler) GetRecords(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	records, err := h.service.GetParseRecords(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "获取记录失败",
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data:    records,
	})
}

type FetchVideoURLRequest struct {
	PhotoID string `json:"photo_id" binding:"required"`
}

func (h *VideoHandler) FetchVideoURL(c *gin.Context) {
	var req FetchVideoURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	videoURL, err := h.service.FetchVideoURL(req.PhotoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "获取视频URL失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data: map[string]string{
			"video_url": videoURL,
		},
	})
}

type FetchAtlasImagesRequest struct {
	PhotoID string `json:"photo_id" binding:"required"`
}

func (h *VideoHandler) FetchAtlasImages(c *gin.Context) {
	var req FetchAtlasImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "请求参数错误: " + err.Error(),
		})
		return
	}

	images, err := h.service.FetchAtlasImages(req.PhotoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "获取图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ParseResponse{
		Success: true,
		Data: map[string]interface{}{
			"images": images,
		},
	})
}

func (h *VideoHandler) ProxyImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	platform, err := utils.ResourcePlatform(imageURL)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if platform == "xhs" {
		req.Header.Set("Referer", "https://www.xiaohongshu.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	} else if platform == "douyin" {
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)")
	} else if platform == "kuaishou" {
		req.Header.Set("Referer", "https://www.kuaishou.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X)")
	} else if platform == "miyoushe" {
		req.Header.Set("Referer", "https://m.miyoushe.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36")
	}

	resp, err := downloadClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Status(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=86400")

	io.Copy(c.Writer, io.LimitReader(resp.Body, maxBodySize))
}

func (h *VideoHandler) Download(c *gin.Context) {
	videoURL := c.Query("url")
	filename := c.Query("filename")
	if videoURL == "" {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "缺少url参数",
		})
		return
	}

	if err := utils.IsAllowedDownloadURL(videoURL); err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "不允许下载此URL: " + err.Error(),
		})
		return
	}

	platform, err := utils.ResourcePlatform(videoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, ParseResponse{
			Success: false,
			Error:   "不支持的资源URL: " + err.Error(),
		})
		return
	}
	isXhsResource := platform == "xhs"

	if isXhsResource {
		req, err := http.NewRequest("GET", videoURL, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ParseResponse{
				Success: false,
				Error:   "创建请求失败: " + err.Error(),
			})
			return
		}
		req.Header.Set("Referer", "https://www.xiaohongshu.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

		resp, err := downloadClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, ParseResponse{
				Success: false,
				Error:   "下载失败",
			})
			return
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Content-Type", contentType)
		if filename == "" {
			filename = "file"
		}
		c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")

		io.Copy(c.Writer, io.LimitReader(resp.Body, maxBodySize))
		return
	}

	if filename == "" {
		filename = "video.mp4"
	}

	req, err := http.NewRequest("GET", videoURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "创建请求失败: " + err.Error(),
		})
		return
	}

	switch {
	case platform == "douyin":
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	case platform == "kuaishou":
		req.Header.Set("Referer", "https://www.kuaishou.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.0 Mobile/15E148 Safari/604.1")
	case platform == "miyoushe":
		req.Header.Set("Referer", "https://m.miyoushe.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Mobile Safari/537.36")
	default:
		req.Header.Set("Referer", "https://www.douyin.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "下载失败: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, ParseResponse{
			Success: false,
			Error:   "远程返回状态码: " + strconv.Itoa(resp.StatusCode),
		})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")

	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		c.Header("Content-Length", contentLength)
	}

	c.Status(http.StatusOK)
	limitedReader := io.LimitReader(resp.Body, maxBodySize)
	io.Copy(c.Writer, limitedReader)
}
