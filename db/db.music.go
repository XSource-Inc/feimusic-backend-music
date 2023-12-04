package db

import (
	"context"
	"strings"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
	"gorm.io/gorm"
)

func AddMusic(ctx context.Context, newMusic *model.Music) error {
	logs.CtxInfo(ctx, "[DB] add music=%v", newMusic)
	err := db.Create(newMusic).Error

	if err != nil {
		logs.CtxWarn(ctx, "failed to add music, err=%v", err)
		return err
	}

	return nil
}

func DeleteMusicWithID(ctx context.Context, tx *gorm.DB, musicID, userID int64) error {
	logs.CtxInfo(ctx, "[DB] delete music=%v", musicID)
	err := tx.Table("music").Where("music_id = ? and user_id = ?", musicID, userID).Update("status", 1).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		return err
	}
	return nil
}

func UpdateMusic(ctx context.Context, musicID int64, updateData map[string]any) error {
	logs.CtxInfo(ctx, "[DB] update music, musid id=%v, data=%v", musicID, updateData)
	var music model.Music
	res := db.Model(&music).Where("music_id = ?", musicID).UpdateColumns(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetMusicWithUniqueMusicID(ctx context.Context, musicID int64) (*model.Music, error) {
	logs.CtxInfo(ctx, "[DB] get music, musid id=%v", musicID)
	var music model.Music
	err := db.First(&music, musicID).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		return nil, err
	}
	return &music, nil
}

func BatchGetMusicWithMsuicID(ctx context.Context, musicIDs []string) ([]*music.MusicItem, error) {
	logs.CtxInfo(ctx, "[DB] batch get music with music id, music ids=%v", musicIDs)
	var songs []model.Music
	var musicList []*music.MusicItem
	err := db.Table("music").Where("music_id in ?", musicIDs).Find(&songs).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		return musicList, err
	}
	var artist, tags []string

	for _, m := range songs {

		var musicItem *music.MusicItem

		artist = strings.Split(m.Artist, ",")
		tags = strings.Split(m.Tags, ",")
		
		musicItem.MusicId = m.MusicID
		musicItem.MusicName = m.MusicName
		musicItem.Artist = artist
		musicItem.Album = m.Album
		musicItem.Tags = tags
		musicItem.UserId = m.UserID
		
		musicList = append(musicList, musicItem)
	}
	return musicList, nil
}

func SearchMusic(ctx context.Context, req *music.SearchMusicRequest) (*[]model.Music, int64, error) {
	query := db.Model(&model.Music{})

	if req.MusicName != nil {
		query = query.Where("music_name = ?", req.MusicName) //TODO：后边要支持模糊查询
	}

	if req.Artist != nil {
		query = query.Where("artist LIKE ?", "%"+*req.Artist+"%") //TODO:用ES才行，后期优化
	}

	if req.UserId != nil {
		query = query.Where("user_id = ?", req.UserId) 
	}

	if req.Tags != nil {
		query = query.Where("tags LIKE ?", "%"+*req.Tags+"%")
	}

	if req.Album != nil {
		query = query.Where("album = ?", req.Album)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logs.CtxWarn(ctx, "failed to search music, err=%v", err)
		return nil, 0, err
	}

	query = query.Limit(int(req.Size)).Offset(int(req.Page * req.Size))

	var music []model.Music
	if err := query.Find(&music).Error; err != nil {
		logs.CtxWarn(ctx, "failed to search music, err=%v", err)
		return nil, 0, err
	}

	return &music, total, nil 
}
