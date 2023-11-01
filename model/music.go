package model

import (
	"time"
)

type Music struct {
	MusicID   string     `gorm:"primary_key"`
	MusicName string     `gorm:"music_name"`
	Artist    string     `gorm:"artist"`
	Album     *string    `gorm:"album"`
	Tags      string     `gorm:"tags"` 
	UserID    string     `gorm:"user_id"`
	Status    int32      `gorm:"status"`
	CreatedAt time.Time  `gorm:"createat"`
	UpdatedAt *time.Time `gorm:"updateat"`
	DeletedAt *time.Time `gorm:"deleteat"`
}

// 增加md5字段
