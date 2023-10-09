package handler

import (
	"context"

	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)

type FeiMusicMusicList struct {
	music.UnimplementedFeiMusicMusicListServer //这个成员是什么
}


// 参数名in更好？
func (ml *FeiMusicMusicList)CreateMusicList(ctx context.Context, in *music.CreateMusicListRequest) (*music.CreateMusicListResponse, error){
	userID := GetValue(ctx, "user_id")
	// 同一个用户的歌单不要重名？
	resp := &music.CreateMusicListRespons{}
	dupl, err:= db.IsDuplicateMusicList(ctx, in.ListName, userID)
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}

	if dupl {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, error.New("歌单名称重复，请修改")
		// new的报错和statusmessage都是给用户看的吗？不重复了吗
	}
	newMusicList := &model.MusicList{
		ListName: in.ListName,
		MusicIDs: []string{},
		ListComment: in.ListComment,
		Tags: in.Tags,
		UserID: userID,
	}

	err = db.CreateMusicList(ctx, newMusicList)
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	return resp, nil
}



func (ml *FeiMusicMusicList)DeleteMusicList(ctx context.Context, in *music.DeleteMusicListRequest) (*music.DeleteMusicListResponse, error){
	// 有业务限制吗，user_id == userID就行吧？
	userID := GetValue(ctx, "user_id")
	user_id, err := db.GetUserIDWithListID(ctx, in.ListID) // 这个适用性好局限，要改进吗
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain user_id of music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	if user_id == ""{
		// 要删的数据的user_id找不到，要报错还是？
		return resp, err
	}
	if user_id != UserID{
		logs.CtxWarn(ctx, "No permission to delete this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, error.New("非本人歌单，没有删除权限")
	}
	// 以上代码别的接口也需要，单独抽出个函数还是做成中间件？
	err = db.DeleteMusicList(ctx, in.MusicID)
	if err != nil{
		logs.CtxWarn(ctx, "failed to delete music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	return resp, nil
}



func (ml *FeiMusicMusicList)UpdateMusicList(ctx context.Context, in *music.UpdateMusicListRequest) (*music.UpdateMusicListResponse, error){
	updateData := map[string]any{}
	utils.AddToMapIfNotNil(updateData, req.ListName)
	utils.AddToMapIfNotNil(updateData, req.ListComment)
	utils.AddToMapIfNotNil(updateData, req.Tags)
	listID := in.ListID
	// 做变更后的唯一性判断
	dupl, err := db. 

	resp := &user.MusicUpdateResponse{}
	if nums == 1 and music_id != req.MusicID{// 这个逻辑处理的对吗// 没有处理查到了多条的异常场景，正常入库不会有多条，要处理吗
		logs.CtxWarn(ctx, "music repeat, the duplicate music id is %v", music_id)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, errors.New("修改后的音乐与音乐库中其他音乐重复")
	} 
	err = db.UpdateMusic(ctx, req.MusicID, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	
	return resp, nil
}

func (ml *FeiMusicMusicList)GetMusicFromList(ctx context.Context, in *music.GetMusicFromListRequest) (*music.GetMusicFromListResponse, error){
	// 鉴权，看请求的歌单是否归属当前登陆人
	resp := &music.GetMusicFromListResponse{}
	music_list, total, err := db.GetMusicFromList(ctx, in.ListID)
	if err != nil{
		logs.CtxWarn(ctx, "failed to get music from music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	resp.MusicItem = music_list
	resp.Total = total
	return resp, nil
} 



func (ml *FeiMusicMusicList)AddMusicToList(ctx context.Context, in *music.AddMusicToListRequest) (*music.AddMusicToListResponse, error){
	//鉴权，看入参中的歌单是否属于当前登陆人
	err := db.AddMusicToList(ctx, in.MusicID, in.ListID)
	if 
}

func (ml *FeiMusicMusicList)RemoveMusicFromList(ctx context.Context, in *music.RemoveMusicFromListRequest) (*music.RemoveMusicFromListResponse, error){
	return 
}






