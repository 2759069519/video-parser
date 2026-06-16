package model

import (
	"time"

	"gorm.io/gorm"
)

// Video 视频信息
type Video struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 视频基本信息
	Platform    string `json:"platform" gorm:"index"` // 平台: kuaishou, douyin, etc.
	VideoID     string `json:"video_id" gorm:"index"` // 平台视频ID
	Title       string `json:"title"`
	Author      string `json:"author"`
	AuthorID    string `json:"author_id"`
	
	// 视频资源
	VideoURL    string `json:"video_url"`
	CoverURL    string `json:"cover_url"`
	Duration    int    `json:"duration"` // 时长(毫秒)
	
	// 统计信息
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	ViewCount    int64 `json:"view_count"`
	
	// 原始数据
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

// Atlas 图文信息
type Atlas struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 图文基本信息
	Platform    string `json:"platform" gorm:"index"`
	AtlasID     string `json:"atlas_id" gorm:"index"` // 平台图文ID
	Title       string `json:"title"`
	Author      string `json:"author"`
	AuthorID    string `json:"author_id"`
	
	// 图片列表 (JSON 数组)
	Images      string `json:"images" gorm:"type:jsonb"`
	
	// 统计信息
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	
	// 原始数据
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

// Profile 用户主页信息
type Profile struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 用户基本信息
	Platform    string `json:"platform" gorm:"index"`
	UserID      string `json:"user_id" gorm:"index"`
	UserName    string `json:"user_name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	
	// 统计信息
	FanCount    int64 `json:"fan_count"`
	FollowCount int64 `json:"follow_count"`
	PhotoCount  int64 `json:"photo_count"`
	
	// 最新作品 (JSON 数组)
	LatestPhotos string `json:"latest_photos" gorm:"type:jsonb"`
	
	// 原始数据
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

// ParseRecord 解析记录
type ParseRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// 解析信息
	Platform string `json:"platform" gorm:"index"`
	URL      string `json:"url"`
	Type     string `json:"type"` // video, atlas, profile
	Status   string `json:"status"` // success, failed
	Error    string `json:"error"`
	
	// 关联ID
	VideoID   *uint `json:"video_id"`
	AtlasID   *uint `json:"atlas_id"`
	ProfileID *uint `json:"profile_id"`
}