package model


import (
	"github.com/jinzhu/gorm" 
	"time"
)

type MusicList struct {
	ListID      string `gorm:"primary_key"`
	ListName    string
	MusicIDs    []string
	ListComment string 
	Tags        []string
	UserID		string
	CreatedAt   time.Time
	UpdatedAt   *time.Time 
	DeletedAt   *time.Time 
}
