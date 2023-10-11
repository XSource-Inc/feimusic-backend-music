package model


import (
	"github.com/jinzhu/gorm" 
	"time"
)

type MusicList struct {
	ListID      string `gorm:"primary_key"` //TODO:这个tag可以省略吗
	ListName    string `grom:"list_name"`
	MusicIDs    []string`grom:"music_ids"`
	ListComment string `grom:"list_comment"`
	Tags        []string`grom:"tags"`
	UserID		string`grom:"user_id"`
	CreatedAt   time.Time`grom:"createat"`
	UpdatedAt   *time.Time `grom:"updateat"`
	DeletedAt   *time.Time `grom:"deleteat"`
}
