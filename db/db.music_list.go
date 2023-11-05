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
	musicList := &model.MusicList{}
	err := db.Table("music_list").Where("list_name = ? and user_id = ?", listName, userID).First(&musicList).Error
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

// 删除歌单
func DeleteMusicList(ctx context.Context, listID string) error {
	logs.CtxInfo(ctx, "[DB] delete music list, list id=%v", listID)
	musicList := model.MusicList{}
	res := db.Model(&musicList).Where("list_id = ?", listID).UpdateColumn(map[string]any{"status": 1})
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

// 删除歌单下音乐
func DeleteListMusic(ctx context.Context, listID string) error {
	logs.CtxInfo(ctx, "[DB] delete music from specified music list, list id=%v", listID)
	ListMusic := model.ListMusic{}
	res := db.Model(&ListMusic).Where("list_id = ?", listID).UpdateColumn(map[string]any{"status": 1}) // 这里的更新，gorm是加了事务的
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music from specified music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func UpdateMusicList(ctx context.Context, listID string, updateData map[string]any) error {
	logs.CtxInfo(ctx, "[DB] update music list, musid list id=%v, data=%v", listID, updateData)
	var musicList model.MusicList
	res := db.Model(&musicList).Where("list_id = ?", listID).UpdateColumns(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

// 获取指定歌单下的音乐id
func GetMusicFromList(ctx context.Context, listID string, status int32) ([]string, error) {
	logs.CtxInfo(ctx, "[DB] get music from music list, list id=%v", listID)
	var Listmusic []string
	var err error
	// 不限制状态时，status传入-1，限制时，status传入指定状态
	if status == -1 {
		err = db.Model(&model.ListMusic{}).Where("list_id = ?", listID).Pluck("music_id", &Listmusic).Error
	} else {
		err = db.Model(&model.ListMusic{}).Where("list_id = ? and status = ?", listID, status).Pluck("music_id", &Listmusic).Error
	}

	if err != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", err)
		return nil, err
	}

	return Listmusic, nil
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
	/*
		// 已删除的音乐, 和需要添加的音乐求交集，更新status为1
		deleteMusicIDs, err := GetMusicFromList(ctx, listID, 1)
		updateMusicIDs := utils.Intersection(deleteMusicIDs, musicIDs)

		err = BatchUpdateMusicStatus(ctx, listID, updateMusicIDs, 0)

		if err != nil {
			logs.CtxWarn(ctx, "failed to append music ID to list, err=%v", err) // TODO:这里的处理不太合理，即使发生错误，也要继续后续的添加操作？
			return err
		}

		// 正常状态的音乐过滤掉已删除的音乐，得到otherMusicIDs
		otherMusicIDs := utils.FilterItem(musicIDs, updateMusicIDs)

		// otherMusicIDs直接添加，包括原来已存在和原来不存在的音乐 ==》这里处理的不对，gorm检测到冲突之后会报错
		var listMusics []model.ListMusic
		for _, musicID := range otherMusicIDs {
			listMusics = append(listMusics, model.ListMusic{ListID: listID, MusicID: musicID, Status: 0})
		}
		err = db.Model(&model.ListMusic{}).Create(&listMusics).Error
	*/
	var listMusics []model.ListMusic
	for _, musicID := range musicIDs {
		listMusics = append(listMusics, model.ListMusic{ListID: listID, MusicID: musicID, Status: 0})
	}
	// 如果数据库中已存在与主键相同的记录，则会更新该记录的其他字段；如果不存在，则插入新记录
	err := db.Model(&model.ListMusic{}).Save(&listMusics).Error // TODO:save是根据主键来判断的，要给歌单和歌曲字段创建个联合主键
	if err != nil {
		logs.CtxWarn(ctx, "failed to append music ID to list, err=%v", err)
		return err
	}

	return nil
}

// 批量修改歌单中歌曲的状态
func BatchUpdateMusicStatus(ctx context.Context, listID string, musicIDs []string, status int32) error {
	logs.CtxInfo(ctx, "[DB] update the status of music in the music list, music id in %v, music list id=%v, status=%v", musicIDs, listID, status)
	ListMusic := model.ListMusic{}
	err := db.Model(&ListMusic).Where("list_id = ? and music_id IN ?", listID, musicIDs).UpdateColumn(map[string]any{"status": status}).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music from music list, err=%v", err)
		return err
	}
	return nil
}

func GetListWithUserID(ctx context.Context, userID string) ([]string, error) {
	logs.CtxInfo(ctx, "[DB] get list id with user id, user id=%v", userID)
	var lists []string
	err := db.Model(&model.MusicList{}).Where("user_id = ?", userID).Pluck("list_id", &lists).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to get list with user id, err=%v", err)
		return lists, err
	}
	return lists, nil
}

func DeleteMusicFromList(ctx context.Context, musicID string,  listID []string) error {
	logs.CtxInfo(ctx, "[DB] delete music from music list, music id=%v, list id=%v", musicID, listID)
	err := db.Model(&model.ListMusic{}).Where("list_id in ? and music_id = ?", listID, musicID).UpdateColumn(map[string]any{"status": 1}).Error
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music from list, err=%v", err)
		return err
	}
	return nil
}
