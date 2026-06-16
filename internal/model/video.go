package model

import (
	"time"

	"gorm.io/gorm"
)

type Video struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	Platform    string `json:"platform" gorm:"index"`
	VideoID     string `json:"video_id" gorm:"index"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	AuthorID    string `json:"author_id"`
	
	VideoURL    string `json:"video_url"`
	CoverURL    string `json:"cover_url"`
	Duration    int    `json:"duration"`
	
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	ViewCount    int64 `json:"view_count"`
	
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

func (Video) TableName() string {
	return "videos"
}

type Atlas struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	Platform    string `json:"platform" gorm:"index"`
	AtlasID     string `json:"atlas_id" gorm:"index"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	AuthorID    string `json:"author_id"`
	
	Images      string `json:"images" gorm:"type:jsonb"`
	
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`
	
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

func (Atlas) TableName() string {
	return "atlases"
}

type Profile struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	Platform    string `json:"platform" gorm:"index"`
	UserID      string `json:"user_id" gorm:"index"`
	UserName    string `json:"user_name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	
	FanCount    int64 `json:"fan_count"`
	FollowCount int64 `json:"follow_count"`
	PhotoCount  int64 `json:"photo_count"`
	
	LatestPhotos string `json:"latest_photos" gorm:"type:jsonb"`
	
	RawData string `json:"raw_data" gorm:"type:jsonb"`
}

func (Profile) TableName() string {
	return "profiles"
}

type ParseRecord struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Platform string `json:"platform" gorm:"index"`
	URL      string `json:"url"`
	Type     string `json:"type" gorm:"index"`
	Status   string `json:"status" gorm:"index"`
	Error    string `json:"error"`
	ErrorCategory string `json:"error_category" gorm:"index"`
	
	Duration    int64  `json:"duration"` // 耗时(毫秒)
	IPAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	
	VideoID   *uint `json:"video_id"`
	AtlasID   *uint `json:"atlas_id"`
	ProfileID *uint `json:"profile_id"`
}

func (ParseRecord) TableName() string {
	return "parse_records"
}
