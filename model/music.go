package model

import (
	"github.com/jinzhu/gorm" 
	"time"
)

type Music struct {
	MusicID   string `gorm:"primary_key"`
	MusicName string
	Artist    []string
	Album     string
	Tags      []string
	UserID    string
	CreatedAt time.Time
	UpdatedAt time.Time // 这个字段有必要吗,加*吗
	DeletedAt *time.Time 
}

// 相关表都写一个文件里吗

