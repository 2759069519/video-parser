package service

import (
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
}

func NewVideoService() *VideoService {
	return &VideoService{
		kuaishouParser:    parser.NewKuaishouParser(),
		douyinParser:      parser.NewDouyinParser(),
		xiaohongshuParser: parser.NewXiaohongshuParser(),
		miyousheParser:    parser.NewMiyousheParser(),
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
	}
	return nil, ""
}

func (s *VideoService) ParseURL(url string) (interface{}, string, error) {
	p, platform := s.getParserByURL(url)
	if p == nil {
		return nil, "", errors.New("不支持的平台")
	}

	result, contentType, err := p.ParseURL(url)
	if err != nil {
		s.recordParse(url, platform, "failed", err.Error(), nil, nil, nil, contentType)
		return nil, "", err
	}

	switch contentType {
	case "video":
		videoResult, ok := result.(*parser.VideoResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for video")
		}
		video, err := s.saveVideoResult(videoResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", &video.ID, nil, nil, contentType)
		return videoResult, contentType, nil
	case "atlas":
		atlasResult, ok := result.(*parser.AtlasResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for atlas")
		}
		atlas, err := s.saveAtlasResult(atlasResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", nil, &atlas.ID, nil, contentType)
		return atlasResult, contentType, nil
	case "profile":
		profileResult, ok := result.(*parser.ProfileResult)
		if !ok {
			s.recordParse(url, platform, "failed", "parser result type mismatch", nil, nil, nil, contentType)
			return nil, "", errors.New("parser result type mismatch for profile")
		}
		profile, err := s.saveProfileResult(profileResult, platform)
		if err != nil {
			s.recordParse(url, platform, "failed", err.Error(), nil, nil, nil, contentType)
			return nil, "", err
		}
		s.recordParse(url, platform, "success", "", nil, nil, &profile.ID, contentType)
		return profileResult, contentType, nil
	}

	return nil, "", errors.New("未知的内容类型")
}

func (s *VideoService) saveVideoResult(result *parser.VideoResult, platform string) (*model.Video, error) {
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

func (s *VideoService) recordParse(url, platform, status, errorMsg string, videoID, atlasID, profileID *uint, contentType string) {
	record := &model.ParseRecord{
		Platform:  platform,
		URL:       url,
		Type:      contentType,
		Status:    status,
		Error:     errorMsg,
		VideoID:   videoID,
		AtlasID:   atlasID,
		ProfileID: profileID,
	}
	if err := repository.DB.Create(record).Error; err != nil {
		log.Printf("记录解析日志失败: %v", err)
	}
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

func (s *VideoService) GetParseRecords(limit int) ([]model.ParseRecord, error) {
	var records []model.ParseRecord
	if err := repository.DB.Order("created_at desc").Limit(limit).Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (s *VideoService) FetchVideoURL(photoID string) (string, error) {
	return s.kuaishouParser.FetchVideoURL(photoID)
}

func (s *VideoService) FetchAtlasImages(photoID string) ([]parser.Image, error) {
	return s.kuaishouParser.FetchAtlasImages(photoID)
}
