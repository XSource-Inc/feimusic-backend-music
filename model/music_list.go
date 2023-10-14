package model

import (
	"time"
)

type MusicList struct {
	ListID      string     `gorm:"primary_key"` // TODO:主键全部改成int64
	ListName    string     `grom:"list_name"`
	MusicIDs    []string   `grom:"music_ids"` // TODO：mysql没有[]string，待修正，方案一改成json串
	ListComment *string    `grom:"list_comment"` // TODO：这里一般和IDL中的定义保持一致？
	Tags        []string   `grom:"tags"`
	UserID      string     `grom:"user_id"`
	CreatedAt   time.Time  `grom:"createat"`
	UpdatedAt   *time.Time `grom:"updateat"`
	DeletedAt   *time.Time `grom:"deleteat"`
}

// TODO：方案二，音乐列表应该设计两张表。一张存音乐列表的基础信息，一张存音乐列表和音乐的关系
