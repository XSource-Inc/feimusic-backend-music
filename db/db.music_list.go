package db

func IsDuplicateMusicList(ctx context.Context, listName, userID string)(bool, error){
	logs.CtxInfo(ctx, "[DB] determine if the song title is duplicated, list name=%v, user=%v", listNmae, userID)
	musicList := &model.MusicList{ListName: listName, userID: userID}
	res := db.First(&musicList)
	if res.Error != nil{
		logs.CtxWarn(ctx, "failed to get music list, err=%v", res.Error)
		return False, res.Error
	}
	if musicList.ListID != ""{
		return True, nil // TODO:这么判断合适吗
	} 
	return False, nil
}


func CreateMusicList(ctx, newMusicList *model.MusicList)(error){
	logs.CtxInfo(ctx, "[DB] create music list, data=%v", newMusicList)
	res := db.Create(newMusicList)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetUserIDWithListID(ctx context.Context, listID string)(string, error){
	logs.CtxInfo(ctx, "[DB] get user id with music list id, list id=%v", ListID)
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

func GetMusicFromList(ctx context.Context, list_id string)(*[]music.MusicList, int64, error){
	music_list := []music.MusicList{}
	total := 0
	res := db.Fetch(&music_list, ListID = list_id)//fetch是这么写吗？
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get music from list, err=%v", res.Error)
		return nil, 0, res.Error
	}
	total = res.//?
	return music_list, total, nil
	//不会写
}