package db

func IsDuplicateMusicList(ctx context.Context, list_name, user_id string)(bool, error){
	music_list := &model.MusicList{ListName: list_name, userID: user_id}
	res := db.First(&music_list)
	if res.Error != nil{
		logs.CtxWarn(ctx, "failed to get music list, err=%v", res.Error)
		return True, res.Error
	}
	if music_list.ListID != ""{
		return True, nil // 这么判断对吗
	} 
	return Falsee, nil
}


func CreateMusicList(ctx, new_music_list *model.MusicList)(error){
	db.Create(new_music_list)
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to add music list, err=%v", res.Error)
		return res.Error
	}
	return nil
}

func GetUserIDWithListID(ctx context.Context, list_id string)(string, error){
	music_list := model.MusicList{ListID: list_id}
	res := db.First(&music_list) // 这样查出来的字段太多了，我只需要一个字段就够了，怎么写
	if res.Error != nil {
		logs.CtxWarn(ctx, "failed to get user_id of music_list, err=%v", res.Error)
		return "", res.Error
	}
	return music_list.userID, nil
}

func DeleteMusicList(ctx context.Context, music_id string)(error){
	music_list := model.MusicList{}
	res := db.Delete(&music_list, music_id) 
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