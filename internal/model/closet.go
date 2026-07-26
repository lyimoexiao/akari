package model

import "time"

// UserCloset records a texture collected by a user (many-to-many with extra item_name).
type UserCloset struct {
	UserID     uint      `gorm:"primaryKey" json:"user_id"`
	TextureTID uint      `gorm:"primaryKey" json:"texture_tid"`
	ItemName   string    `gorm:"size:128;not null" json:"item_name"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Associations (loaded on demand)
	Texture *Texture `gorm:"foreignKey:TextureTID;references:TID" json:"texture,omitempty"`
}

func (UserCloset) TableName() string {
	return "user_closet"
}
