package db

import (
	"context"
	"errors"
	"strings"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)

func JudgeMusicWithUniqueNameAndArtist(ctx context.Context, musicName, musicArtist string) error {
	logs.CtxInfo(ctx, "[DB] determine the uniqueness of a song based on song name and artist, song name=%v, artist=%v", musicName, musicArtist)
	music := model.Music{
		MusicName: musicName,
		Artist:    musicArtist,
	}

	err := db.First(&music).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		return err
	}

	return nil
}

// func JudgeMusicWithMusicID(ctx context.Context, musicID string)(bool, error){
// 	logs.CtxInfo(ctx, "[DB] check if the music to be deteled exist, music id=%v", musicID)
// 	music := model.Music{
// 		MusicID: musicID,
// 	}

// 	err := db.Frist(&music).Error
// 	if err != nil{
// 		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
// 		return false, err
// 	}
// 	return true, nil
// }

func AddMusic(ctx context.Context, newMusic *model.Music) error {
	logs.CtxInfo(ctx, "[DB] add music=%v", newMusic)
	err := db.Create(newMusic).Error

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			logs.CtxWarn(ctx, "the music library already contains this singer's music, singer=%v, music name=%v", newMusic.Artist, newMusic.MusicName)
			return errors.New("duplicate entry")
		} else {
			logs.CtxWarn(ctx, "failed to add music, err=%v", err)
			return err
		}
	}
	return nil
}

func DeleteMusicWithID(ctx context.Context, musicID, userID string) error {
	logs.CtxInfo(ctx, "[DB] delete music=%v", musicID)
	err := db.Table("music").Where("music_id = ? and user_id = ?", musicID, userID).UpdateColumn(map[string]any{"status": 1}).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		return err
	}
	return nil
}

func UpdateMusic(ctx context.Context, musicID string, updateData map[string]any) error {
	logs.CtxInfo(ctx, "[DB] update music, musid id=%v, data=%v", musicID, updateData)
	var music model.Music
	res := db.Model(&music).Where("music_id = ?", musicID).UpdateColumns(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func SearchMusic(ctx context.Context, req *music.SearchMusicRequest) (*[]model.Music, int64, error) { //加*吗
	return nil, 0, nil
}

func GetMusicWithUniqueMusicID(ctx context.Context, musicID string) (*model.Music, error) {
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
		artist = strings.Split(m.Artist, ",")
		tags = strings.Split(m.Tags, ",")
		var musicItem *music.MusicItem
		musicItem.MusicId = m.MusicID
		musicItem.MusicName = m.MusicName
		musicItem.Artist = artist
		musicItem.Album = *m.Album
		musicItem.Tags = tags
		musicItem.UserId = m.UserID
		musicList = append(musicList, musicItem)
	}
	return musicList, nil
}
