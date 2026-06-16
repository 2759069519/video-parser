package service

import (
	"time"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"video-parser/internal/model"
	"video-parser/internal/parser"
	"video-parser/internal/repository"
)

type VideoService struct {
	kuaishouParser    *parser.KuaishouParser
	douyinParser      *parser.DouyinParser
	xiaohongshuParser *parser.XiaohongshuParser
	miyousheParser    *parser.MiyousheParser
	doubaoParser      *parser.DoubaoParser
}

func NewVideoService() *VideoService {
	return &VideoService{
		kuaishouParser:    parser.NewKuaishouParser(),
		douyinParser:      parser.NewDouyinParser(),
		xiaohongshuParser: parser.NewXiaohongshuParser(),
		miyousheParser:    parser.NewMiyousheParser(),
		doubaoParser:      parser.NewDoubaoParser(),
	}
}

func (s *VideoService) getParserByURL(url string) (parser.Parser, string) {
	platform := parser.DetectPlatform(url)
	switch platform {
	case "kuaishou":
		return s.kuaishouParser, "kuaishou"
	case "douyin":
		return s.douyinParser, "douyin"
	case "xhs":
		return s.xiaohongshuParser, "xhs"
	case "miyoushe":
		return s.miyousheParser, "miyoushe"
	case "doubao":
		return s.doubaoParser, "doubao"
	}
	return nil, ""
}

func (s *VideoService) ParseURL(url string, ipAddress, userAgent string) (interface{}, string, error) {
	start := time.Now()
	p, platform := s.getParserByURL(url)
	if p == nil {
		s.recordParse(url, platform, "failed", "不支持的平台", "platform_not_supported", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, "")
		return nil, "", errors.New("不支持的平台")
	}

	result, contentType, err := p.ParseURL(url)
	if err != nil {
		errCategory := categorizeError(err)
		s.recordParse(url, platform, "failed", err.Error(), errCategory, time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
		return nil, "", err
	}

	switch contentType {
	case "video":
		videoResult, ok := result.(*parser.VideoResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", "internal_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for video")
		}
		video, err := s.saveVideoResult(videoResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), "database_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", "", time.Since(start).Milliseconds(), ipAddress, userAgent, &video.ID, nil, nil, contentType)
		return videoResult, contentType, nil
	case "atlas":
		atlasResult, ok := result.(*parser.AtlasResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", "internal_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for atlas")
		}
		atlas, err := s.saveAtlasResult(atlasResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), "database_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", "", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, &atlas.ID, nil, contentType)
		return atlasResult, contentType, nil
	case "profile":
		profileResult, ok := result.(*parser.ProfileResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", "internal_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for profile")
		}
		profile, err := s.saveProfileResult(profileResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), "database_error", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", "", time.Since(start).Milliseconds(), ipAddress, userAgent, nil, nil, &profile.ID, contentType)
		return profileResult, contentType, nil
	}

	return nil, "", errors.New("未知的内容类型")
}

func categorizeError(err error) string {
	errMsg := err.Error()
	switch {
	case contains(errMsg, "不支持的平台", "资源域名不在允许列表"):
		return "platform_not_supported"
	case contains(errMsg, "超时", "timeout", "deadline exceeded"):
		return "timeout"
	case contains(errMsg, "连接", "connection refused", "dial tcp"):
		return "connection_error"
	case contains(errMsg, "状态码", "status code"):
		return "http_error"
	case contains(errMsg, "解析", "parse", "json"):
		return "parse_error"
	case contains(errMsg, "未找到", "not found"):
		return "not_found"
	default:
		return "unknown"
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

func (s *VideoService) saveVideoResult(result *parser.VideoResult, platform string) (*model.Video, error) {
	var existing model.Video
	if err := repository.DB.Where("platform = ? AND video_id = ?", platform, result.VideoID).First(&existing).Error; err == nil {
		existing.Title = result.Title
		existing.Author = result.Author
		existing.VideoURL = result.VideoURL
		existing.CoverURL = result.CoverURL
		existing.Duration = result.Duration
		existing.LikeCount = result.LikeCount
		existing.CommentCount = result.CommentCount
		existing.ViewCount = result.ViewCount
		rawData, _ := json.Marshal(result)
		existing.RawData = string(rawData)
		if err := repository.DB.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update video: %w", err)
		}
		return &existing, nil
	}

	rawData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal video result: %w", err)
	}

	video := &model.Video{
		Platform:     platform,
		VideoID:      result.VideoID,
		Title:        result.Title,
		Author:       result.Author,
		AuthorID:     result.AuthorID,
		VideoURL:     result.VideoURL,
		CoverURL:     result.CoverURL,
		Duration:     result.Duration,
		LikeCount:    result.LikeCount,
		CommentCount: result.CommentCount,
		ViewCount:    result.ViewCount,
		RawData:      string(rawData),
	}

	if err := repository.DB.Create(video).Error; err != nil {
		return nil, fmt.Errorf("save video: %w", err)
	}
	return video, nil
}

func (s *VideoService) saveAtlasResult(result *parser.AtlasResult, platform string) (*model.Atlas, error) {
	var existing model.Atlas
	if err := repository.DB.Where("platform = ? AND atlas_id = ?", platform, result.AtlasID).First(&existing).Error; err == nil {
		existing.Title = result.Title
		existing.Author = result.Author
		images, _ := json.Marshal(result.Images)
		existing.Images = string(images)
		existing.LikeCount = result.LikeCount
		existing.CommentCount = result.CommentCount
		rawData, _ := json.Marshal(result)
		existing.RawData = string(rawData)
		if err := repository.DB.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update atlas: %w", err)
		}
		return &existing, nil
	}

	images, err := json.Marshal(result.Images)
	if err != nil {
		return nil, fmt.Errorf("marshal images: %w", err)
	}

	rawData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal atlas result: %w", err)
	}

	atlas := &model.Atlas{
		Platform:     platform,
		AtlasID:      result.AtlasID,
		Title:        result.Title,
		Author:       result.Author,
		AuthorID:     result.AuthorID,
		LikeCount:    result.LikeCount,
		CommentCount: result.CommentCount,
		Images:       string(images),
		RawData:      string(rawData),
	}

	if err := repository.DB.Create(atlas).Error; err != nil {
		return nil, fmt.Errorf("save atlas: %w", err)
	}
	return atlas, nil
}

func (s *VideoService) saveProfileResult(result *parser.ProfileResult, platform string) (*model.Profile, error) {
	var existing model.Profile
	if err := repository.DB.Where("platform = ? AND user_id = ?", platform, result.UserID).First(&existing).Error; err == nil {
		existing.UserName = result.UserName
		existing.Avatar = result.Avatar
		existing.Description = result.Description
		existing.FanCount = result.FanCount
		existing.FollowCount = result.FollowCount
		existing.PhotoCount = result.PhotoCount
		photos, _ := json.Marshal(result.Photos)
		existing.LatestPhotos = string(photos)
		rawData, _ := json.Marshal(result)
		existing.RawData = string(rawData)
		if err := repository.DB.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update profile: %w", err)
		}
		return &existing, nil
	}

	photos, err := json.Marshal(result.Photos)
	if err != nil {
		return nil, fmt.Errorf("marshal photos: %w", err)
	}

	rawData, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal profile result: %w", err)
	}

	profile := &model.Profile{
		Platform:     platform,
		UserID:       result.UserID,
		UserName:     result.UserName,
		Avatar:       result.Avatar,
		Description:  result.Description,
		FanCount:     result.FanCount,
		FollowCount:  result.FollowCount,
		PhotoCount:   result.PhotoCount,
		LatestPhotos: string(photos),
		RawData:      string(rawData),
	}

	if err := repository.DB.Create(profile).Error; err != nil {
		return nil, fmt.Errorf("save profile: %w", err)
	}
	return profile, nil
}

func (s *VideoService) recordParse(url, platform, status, errorMsg, errorCategory string, duration int64, ipAddress, userAgent string, videoID, atlasID, profileID *uint, contentType string) {
	record := &model.ParseRecord{
		Platform:      platform,
		URL:           url,
		Type:          contentType,
		Status:        status,
		Error:         errorMsg,
		ErrorCategory: errorCategory,
		Duration:      duration,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		VideoID:       videoID,
		AtlasID:       atlasID,
		ProfileID:     profileID,
	}
	if err := repository.DB.Create(record).Error; err != nil {
		log.Printf("记录解析日志失败: %v", err)
	}
}

type ParseRecordsResult struct {
	Records  []model.ParseRecord `json:"records"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

func (s *VideoService) GetParseRecords(page, pageSize int, platform, status, contentType string) (*ParseRecordsResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := repository.DB.Model(&model.ParseRecord{})
	
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if contentType != "" {
		query = query.Where("type = ?", contentType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.ParseRecord
	offset := (page - 1) * pageSize
	if err := query.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	return &ParseRecordsResult{
		Records:  records,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *VideoService) GetVideoByID(id uint) (*model.Video, error) {
	var video model.Video
	if err := repository.DB.First(&video, id).Error; err != nil {
		return nil, err
	}
	return &video, nil
}

func (s *VideoService) GetAtlasByID(id uint) (*model.Atlas, error) {
	var atlas model.Atlas
	if err := repository.DB.First(&atlas, id).Error; err != nil {
		return nil, err
	}
	return &atlas, nil
}

func (s *VideoService) GetProfileByID(id uint) (*model.Profile, error) {
	var profile model.Profile
	if err := repository.DB.First(&profile, id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *VideoService) FetchVideoURL(photoID, platform string) (string, error) {
	switch platform {
	case "kuaishou":
		return s.kuaishouParser.FetchVideoURL(photoID)
	default:
		return "", fmt.Errorf("平台 %s 暂不支持获取视频地址", platform)
	}
}

func (s *VideoService) FetchAtlasImages(photoID, platform string) ([]parser.Image, error) {
	switch platform {
	case "kuaishou":
		return s.kuaishouParser.FetchAtlasImages(photoID)
	default:
		return nil, fmt.Errorf("平台 %s 暂不支持获取图集图片", platform)
	}
}
