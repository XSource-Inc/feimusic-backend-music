package db

func IsDuplicateMusicList(ctx context.Context, listName, userID string)(bool, string, error){
	logs.CtxInfo(ctx, "[DB] determine if the song title is duplicated, list name=%v, user=%v", listName, userID)
	musicList := &model.MusicList{ListName: listName, userID: userID}
	res := db.First(&musicList)
	if res.Error != nil{
		logs.CtxWarn(ctx, "failed to get music list, err=%v", res.Error)
		return false, "" , res.Error
	}
	if musicList.ListID != ""{
		return true, musicList.ListID, nil // TODO:这么判断合适吗
	} 
	return false, "", nil
}


func CreateMusicList(ctx context.Context, newMusicList *model.MusicList)(error){
	logs.CtxInfo(ctx, "[DB] create music list, data=%v", newMusicList)
	res := db.Create(newMusicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetUserIDWithListID(ctx context.Context, listID string)(string, error){
	logs.CtxInfo(ctx, "[DB] get user id with music list id, list id=%v", listID)
	musicList := model.MusicList{ListID: listID}
	res := db.First(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get user id of music list, err=%v", res.Error)
		return "", res.Error
	}
	return musicList.userID, nil
}

func DeleteMusicList(ctx context.Context, musicID string)(error){
	logs.CtxInfo(ctx, "[DB] delete music list, music id=%v", musicID)
	musicList := model.MusicList{}
	res := db.Delete(&musicList, musicID) 
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func UpdateMusicList(ctx context.Context, listID string, updateData map[string]any)(error){
	logs.CtxInfo(ctx, "[DB] update music list, musid list id=%v, data=%v", listID, updateData)
	var musicList model.MusicList
	res := db.Model(&musicList).Where("list_id = ?", listID).Updates(updateData)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetMusicFromList(ctx context.Context, listID string)(*[]music.MusicList, int64, error){
	logs.CtxInfo(ctx, "[DB] get music from music list, musid list id=%v", listID)
	musicList := []music.MusicList{}
	var total int64

	res := db.Model(&musicList).Where("ListID = ?", listID).Find(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", res.Error)
		return nil, 0, res.Error
	}

	total = res.RowsAffected

	return &musicList, total, nil
}

func JudgeMusicListWithListID(ctx context.Context, listID string) (error) {
	logs.CtxInfo(ctx, "[DB] determine if the playlist exists,list id=%v", ListID)
    musicList := model.MusicList{}
    res := db.First(&musicList, "list_id = ?", listID)
    if res.Error != nil {
        logs.CtxWarn(ctx, "failed to get music list, err=%v", res.Error)
        return res.Error
    }
	return nil
}

func AddMusicToList(ctx context.Context, listID, musicID string)(error) {
	logs.CtxInfo(ctx, "[DB] add music to music list, list id=%v, music id=%v", ListID, musicID)
	// TODO:这里的处理不知道是不是正确
	var musicList model.MusicList
	musicList.MusicIDs = append(musicList.MusicIDs, musicID)
	data := map[string][]string{
		"music_ids": musicList.MusicIDs
	}
	res := db.Model(&musicList).Where("list_id = ?", listID).Updates(data)
    if res.Error != nil {
        logs.CtxWarn(ctx, "failed to append music ID to list, err=%v", res.Error)
        return res.Error
    }

    return nil
}


func DeleteMusicWithID(ctx context.Context, listID, musicID string)(error){
	logs.CtxInfo(ctx, "[DB] delete music from playlist, music id=%v, music list id=%v", musicID, listID)
	var musicList model.MusicList
	res := db.Model(&musicList).Where("ListID = ?", listID).Find(&musicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", res.Error)
		return res.Error
	}

	musicList.MusicIDs = utils.removeString(musicList.MusicIDs, musicID)
	data := map[string][]string{
		"music_ids": musicList.MusicIDs
	}
	res := db.Model(&musicList).Where("list_id = ?", listID).Updates(data)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to delete music from music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}