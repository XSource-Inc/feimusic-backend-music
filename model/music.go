package model

import (
	"time"
)

type Music struct {
	MusicID   int64     `gorm:"primary_key"`
	MusicName string    `gorm:"music_name"`
	Artist    string    `gorm:"artist"`
	Album     string    `gorm:"album"`
	Tags      string    `gorm:"tags"`
	UserID    int64     `gorm:"user_id"`
	MD5       string    `gorm:"md5"` // 唯一索引
	Status    int16     `gorm:"status"`
	CreatedAt time.Time `gorm:"createat"`
	UpdatedAt time.Time `gorm:"updateat"`
	DeletedAt time.Time `gorm:"deleteat"`
}
