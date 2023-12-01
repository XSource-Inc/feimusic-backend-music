package model

import (
	"time"
)

type MusicList struct {
	ListID      int64     `gorm:"primary_key"`
	ListName    string    `grom:"list_name"`
	ListComment *string   `grom:"list_comment"`
	Tags        string    `grom:"tags"`
	UserID      int64    `grom:"user_id"`
	Status      int32     `gorm:"status"`
	CreatedAt   time.Time `grom:"createat"`
	UpdatedAt   time.Time `grom:"updateat"`
	DeletedAt   time.Time `grom:"deleteat"`
}
// ListName+UserId, 需要一个唯一约束

type ListMusic struct {
	ListID    int64    `gorm:"primary_key"`
	MusicID   int64    `gorm:"music_id"`
	Status    int32     `gorm:"status"`
	CreatedAt time.Time `grom:"createat"`
	UpdatedAt time.Time `grom:"updateat"`
	DeletedAt time.Time `grom:"deleteat"`
}
