package model

import (
	"time"
)

type MusicList struct {
	ListID      string     `gorm:"primary_key"` // TODO:主键全部改成int64
	ListName    string     `grom:"list_name"`
	ListComment *string    `grom:"list_comment"` // TODO：这里一般和IDL中的定义保持一致？
	Tags        string     `grom:"tags"`
	UserID      string     `grom:"user_id"`
	Status      int32      `gorm:"status"`
	CreatedAt   time.Time  `grom:"createat"`
	UpdatedAt   *time.Time `grom:"updateat"`
	DeletedAt   *time.Time `grom:"deleteat"`
}

type ListMusic struct {
	ListID    string     `gorm:"primary_key"`
	MusicID   string     `gorm:"music_id"`
	Status    int32      `gorm:"status"`
	CreatedAt time.Time  `grom:"createat"`
	UpdatedAt *time.Time `grom:"updateat"`
	DeletedAt *time.Time `grom:"deleteat"`
}
