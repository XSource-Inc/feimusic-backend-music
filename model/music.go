package model

import (
	"github.com/jinzhu/gorm" 
	"time"
)

type Music struct {
	MusicID   string `gorm:"primary_key"`
	MusicName string `gorm:"music_name"`
	Artist    []string `gorm:"artist"`
	Album     string `gorm:"album"`
	Tags      []string `gorm:"tags"`
	UserID    string `gorm:"user_id"`
	Status    int32 `gorm:"status"`
	CreatedAt time.Time `gorm:"createat"`
	UpdatedAt time.Time `gorm:"updateat"`// 这个字段有必要吗,加*吗
	DeletedAt *time.Time `gorm:"deleteat"`
}

// 相关表都写一个文件里吗

