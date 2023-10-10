package db

import (
	"context"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/jinzhu/gorm"
)

func JudgeMusicWithUniqueNameAndArtist(ctx context.Context, musicName string, musicArtist []string)(error){
	logs.CtxInfo(ctx, "[DB] determine the uniqueness of a song based on song name and artist, song name=%v, artist=%v", musicName, musicArtist)
	music := model.Music{
		MusicName: musicName,
		Artist: musicArtist,
	}

	err := db.First(&music).Error
	if err != nil{
		logs.CtxWarn(ctx, "failed to get user, err=%v", err)
		return nil, err
	}

	return nil
}

func AddMusic(ctx context.Context, newMusic *model.Music)(error){
	logs.CtxInfo(ctx, "[DB] add music=%v", newMusic)
	res := db.Create(newMusic)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to add music, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func DeleteMusicWithID(ctx context.Context, musicID string)(error){
	logs.CtxInfo(ctx, "[DB] delete music=%v", musicID)
	music := model.Music{}
	res := db.Delete(&music, musicID) 
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func UpdateMusic(ctx context.Context, musicID string, updateData map[string]any)(error){
	logs.CtxInfo(ctx, "[DB] update music, musid id=%v, data=%v", musicID, updateData)
	var music model.Music
	res := db.Model(&music).Where("music_id = ?", musicID).Updates(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func db.SearchMusic(ctx context.Context, req *music.SearchMusicRequest)(*[]model.Music, int64, error){//加*吗
	return nil, 0, nil
}

func db.GetMusicWithUniqueMusicID(ctx context.Context, musicID string)(*model.Music, error){
	var music model.Music
	res := db.First(&music, musicID)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", res.Error)
		return nil, res.Error
	}
	return music, nil
}