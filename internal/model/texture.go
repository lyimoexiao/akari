package model

import (
	"time"

	"gorm.io/gorm"
)

// Texture represents an uploaded skin or cape texture.
type Texture struct {
	TID       uint           `gorm:"primaryKey;autoIncrement" json:"tid"`
	Name      string         `gorm:"size:128;not null" json:"name"`
	Type      string         `gorm:"size:16;not null;index" json:"type"` // "steve", "alex", "cape"
	Hash      string         `gorm:"index;size:128;not null" json:"hash"`
	URL       string         `gorm:"size:512" json:"url"` // full URL to the texture file
	Size      int64          `gorm:"not null" json:"size"`
	Uploader  uint           `gorm:"index;not null" json:"uploader"`
	Public    bool           `gorm:"not null;default:true;index" json:"public"`
	Likes     int64          `gorm:"not null;default:0" json:"likes"`
	UploadAt  time.Time      `gorm:"autoCreateTime" json:"upload_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (Texture) TableName() string {
	return "textures"
}
