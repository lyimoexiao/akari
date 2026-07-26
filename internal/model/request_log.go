package model

import "time"

type RequestLog struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	RequestID       string    `gorm:"index;size:128;not null" json:"request_id"`
	UserID          *uint     `gorm:"index" json:"user_id,omitempty"`
	Module          string    `gorm:"index;size:64;not null" json:"module"`
	Method          string    `gorm:"size:16;not null" json:"method"`
	Path            string    `gorm:"size:2048;not null" json:"path"`
	Status          int       `gorm:"not null" json:"status"`
	LatencyMS       int64     `gorm:"not null" json:"latency_ms"`
	IP              string    `gorm:"size:64;not null" json:"ip"`
	UserAgent       string    `gorm:"size:512;not null" json:"user_agent"`
	RequestHeaders  string    `gorm:"type:text;not null" json:"request_headers"`
	RequestBody     string    `gorm:"type:text;not null" json:"request_body"`
	ResponseHeaders string    `gorm:"type:text;not null" json:"response_headers"`
	ResponseBody    string    `gorm:"type:text;not null" json:"response_body"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}
