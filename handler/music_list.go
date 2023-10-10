package handler

import (
	"context"

	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)

type FeiMusicMusicList struct {
	music.UnimplementedFeiMusicMusicListServer 
}



func (ml *FeiMusicMusicList)CreateMusicList(ctx context.Context, in *music.CreateMusicListRequest) (*music.CreateMusicListResponse, error){
	userID := utils.GetValue(ctx, "user_id")

	resp := &music.CreateMusicListResponse{}
	// TODO:代码结构调整
	dupl, _, err := db.IsDuplicateMusicList(ctx, in.ListName, userID)
	if err != nil and err != gorm.ErrRecordNotFound{ // TODO：这里的判断和处理合适吗？一旦判重失败就不再创建？
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, err
	}

	if dupl {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单名称重复，请修改"}
		return resp, nil
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
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, nil
	}
	return resp, nil
}

func (ml *FeiMusicMusicList)DeleteMusicList(ctx context.Context, in *music.DeleteMusicListRequest) (*music.DeleteMusicListResponse, error){
	// 删除歌单时仅限制操作人是歌单归属人
	resp := &music.DeleteMusicListResponse{}
	// TODO:判断是否有操作权限的代码别的接口也需要，单独抽出个函数还是做成中间件？
	// 定制化的报错怎么处理呢
	userID := utils.GetValue(ctx, "user_id") // TODO:需要考虑取不到userid的情况吗
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListID)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain user_id of music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		// TODO：要考虑这种情况吗，出现这种情况一定是因为表中的值为空字符串？
		logs.CtxWarn(ctx, "failed to obtain user_id of music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}

	if userIDFromTable != UserID{
		logs.CtxWarn(ctx, "No permission to delete this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有删除权限"}
		return resp, nil
	}

	err = db.DeleteMusicList(ctx, in.MusicID)
	if err != nil{
		logs.CtxWarn(ctx, "failed to delete music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}
	return resp, nil
}



func (ml *FeiMusicMusicList)UpdateMusicList(ctx context.Context, in *music.UpdateMusicListRequest) (*music.UpdateMusicListResponse, error){
	resp := &user.MusicUpdateResponse{}
	// 仅可更新本人歌单  TODO:直接获取全部当前用户的歌单，判断要修改的歌单是否在其中是不是更好？
	userID := utils.GetValue(ctx, "user_id") // TODO:需要考虑取不到userid的情况吗
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListID)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain user_id of music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		// TODO：要考虑这种情况吗，出现这种情况一定是因为表中的值为空字符串？
		logs.CtxWarn(ctx, "failed to obtain user_id of music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
		return resp, nil
	}

	if userIDFromTable != UserID{
		logs.CtxWarn(ctx, "No permission to update this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有更新权限"}
		return resp, nil
	}

	listID := in.ListID
	// 做变更后的唯一性判断
	if in.ListName != nil {
		dupl, musicListID, err := db.IsDuplicateMusicList(ctx, in.ListName, userID)
		if err != nil and err != gorm.ErrRecordNotFound{ // TODO：这里的判断和处理合适吗？一旦判重失败就不再创建？
			logs.CtxWarn(ctx, "failed to update music list, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
			return resp, err
		}
	
		if dupl and musicListID != listID {
			logs.CtxWarn(ctx, "failed to update music list, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "修改后的歌单名与已有歌单名重复，请修改"}
			return resp, nil
		}
	}

	updateData := map[string]any{}
	utils.AddToMapIfNotNil(updateData, in.ListName)
	utils.AddToMapIfNotNil(updateData, in.ListComment)
	utils.AddToMapIfNotNil(updateData, in.Tags)

	err = db.UpdateMusicList(ctx, listID, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单信息更新失败"}
		return resp, nil
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






