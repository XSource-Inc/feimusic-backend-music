package db

import (
	"context"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/feimusic-backend-music/utils"
	"github.com/jinzhu/gorm"
)

func IsDuplicateMusicList(ctx context.Context, listName, userID string) (bool, string, error) {
	logs.CtxInfo(ctx, "[DB] determine if the song title is duplicated, list name=%v, user=%v", listName, userID)
	musicList := &model.MusicList{ListName: listName, UserID: userID}
	err := db.First(&musicList).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "", nil
		} else {
			logs.CtxWarn(ctx, "failed to get music list, err=%v", err)
			return false, "", err
		}
	}

	if musicList.ListID != "" {
		return true, musicList.ListID, nil 
	}
	return false, "", nil
}

func CreateMusicList(ctx context.Context, newMusicList *model.MusicList) (string, error) {
	logs.CtxInfo(ctx, "[DB] create music list, data=%v", newMusicList)
	err := db.Create(newMusicList).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		return "", err 
	}
	return newMusicList.ListID, nil
}

func GetUserIDWithListID(ctx context.Context, listID string) (string, error) {
	logs.CtxInfo(ctx, "[DB] get user id with music list id, list id=%v", listID)
	musicList := model.MusicList{ListID: listID}
	res := db.First(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get user id of music list, err=%v", res.Error)
		return "", res.Error
	}
	return musicList.UserID, nil
}

func DeleteMusicList(ctx context.Context, listID string) error {
	logs.CtxInfo(ctx, "[DB] delete music list, list id=%v", listID)
	musicList := model.MusicList{}
	res := db.Model(&musicList).Where("list_id = ?", listID).Update(map[string]any{"status": 1})
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func UpdateMusicList(ctx context.Context, listID string, updateData map[string]any) error {
	logs.CtxInfo(ctx, "[DB] update music list, musid list id=%v, data=%v", listID, updateData)
	var musicList model.MusicList
	res := db.Model(&musicList).Where("list_id = ?", listID).Updates(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetMusicFromList(ctx context.Context, listID string) ([]string, error) {
	logs.CtxInfo(ctx, "[DB] get music from music list, list id=%v", listID)
	musicList := model.MusicList{ListID: listID}

	res := db.First(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", res.Error)
		return nil, res.Error
	}

	return musicList.MusicIDs, nil
}

func FilterMusicIDUsingIDAndStatus(ctx context.Context, musicIDs []string) ([]string, []string, error) {
	logs.CtxInfo(ctx, "[DB] filter music using music id and status, ids=%v", musicIDs)
	var effectiveMusicIDs []string
	var invalidMusicIDS []string
	err := db.Model(&model.Music{}).Where("music_id in ? and status = 0", musicIDs).Pluck("music_id", &effectiveMusicIDs).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to filter music list, err=%v", err)
		return effectiveMusicIDs, invalidMusicIDS, err
	}
	invalidMusicIDS = utils.FilterItem(musicIDs, effectiveMusicIDs)
	return effectiveMusicIDs, invalidMusicIDS, nil
}

func JudgeMusicListWithListID(ctx context.Context, listID string) error {
	logs.CtxInfo(ctx, "[DB] determine if the playlist exists,list id=%v", listID)
	musicList := model.MusicList{}
	res := db.First(&musicList, "list_id = ?", listID)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func AddMusicToList(ctx context.Context, listID string, musicIDs []string) error {
	logs.CtxInfo(ctx, "[DB] add music to music list, list id=%v, music id=%v", listID, musicIDs)
	// 过滤出未添加过的音乐
	listMusicIDs, err := GetMusicFromList(ctx, listID)
	if err != nil {
		logs.CtxWarn(ctx, "failed to append music ID to list, err=%v", err)
		return err
	}
	validMusicIDs := utils.FilterItem(musicIDs, listMusicIDs)



	var musicList model.MusicList
	for _, musicID := range validMusicIDs {
		musicList.MusicIDs = append(musicList.MusicIDs, musicID)
	}
	
	data := map[string][]string{
		"music_ids": musicList.MusicIDs,
	}
	res := db.Model(&musicList).Where("list_id = ?", listID).Update(data)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to append music ID to list, err=%v", res.Error)
		return res.Error
	}

	return nil
}

func DeleteMusicFromList(ctx context.Context, listID, musicID string) error {
	logs.CtxInfo(ctx, "[DB] delete music from playlist, music id=%v, music list id=%v", musicID, listID)
	var musicList model.MusicList
	res := db.Model(&musicList).Where("ListID = ?", listID).Find(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", res.Error)
		return res.Error
	}

	musicList.MusicIDs = utils.RemoveString(musicList.MusicIDs, musicID)
	data := map[string][]string{
		"music_ids": musicList.MusicIDs,
	}
	res = db.Model(&musicList).Where("list_id = ?", listID).Updates(data)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music from music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}
