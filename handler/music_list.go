package handler

import (
	"context"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/db"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/feimusic-backend-music/utils"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/base"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
	"github.com/jinzhu/gorm"
)

type FeiMusicMusicList struct {
	music.UnimplementedFeiMusicMusicListServer // TODO:这里为什么报错了,这里没搞懂结构体包含这个成员的作用
}

func (ml *FeiMusicMusicList)CreateMusicList(ctx context.Context, in *music.CreateMusicListRequest) (*music.CreateMusicListResponse, error){
	userID := utils.GetValue(ctx, "user_id")

	resp := &music.CreateMusicListResponse{}
	// TODO:代码结构调整
	dupl, _, err := db.IsDuplicateMusicList(ctx, in.ListName, userID)
	if err != nil && err != gorm.ErrRecordNotFound{ // 
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

	listID, err := db.CreateMusicList(ctx, newMusicList)
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, nil
	}
	if listID == ""{ // TODO:是否有必要保留？
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err) // TODO：报错信息需要进一步明确吗
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, nil
	}
	resp.ListId = listID
	return resp, nil
}

func (ml *FeiMusicMusicList)DeleteMusicList(ctx context.Context, in *music.DeleteMusicListRequest) (*music.DeleteMusicListResponse, error){
	// 删除歌单时仅限制操作人是歌单归属人
	resp := &music.DeleteMusicListResponse{}

	userID := utils.GetValue(ctx, "user_id") // TODO:需要考虑取不到userid的情况吗
	// TODO：userid应该从in里传过来=》错了吧？
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to delete music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要删除的歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to delete music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
			return resp, nil
		}
	}
	if userIDFromTable == "" {
		// TODO：要考虑这种情况吗，出现这种情况一定是因为表中的值为空字符串？
		// TODO：困惑，是不是只考虑查询结果整体是nil的情况，不考虑业务字段为空的情况？
		// 假入整个记录为nil，会在GetUserIDWithListID，return的musicList.UserID这里触发错误从而被上面捕获？
		logs.CtxWarn(ctx, "failed to delete music list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}

	if userIDFromTable != userID{
		logs.CtxWarn(ctx, "No permission to delete this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有删除权限"}
		return resp, nil
	}
	// 软删除
	err = db.DeleteMusicList(ctx, in.ListId)
	if err != nil{
		logs.CtxWarn(ctx, "failed to delete music list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}
	return resp, nil
}

func (ml *FeiMusicMusicList)UpdateMusicList(ctx context.Context, in *music.UpdateMusicListRequest) (*music.UpdateMusicListResponse, error){
	resp := &music.UpdateMusicListResponse{}
	// TODO:判断要更新的歌单是否存在
	// 仅可更新本人歌单  TODO:直接获取全部当前用户的歌单，判断要修改的歌单是否在其中是不是更好？
	userID := utils.GetValue(ctx, "user_id") // TODO:需要考虑取不到userid的情况吗
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		// TODO：要考虑这种情况吗，出现这种情况一定是因为表中的值为空字符串？
		logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
		return resp, nil
	}

	if userIDFromTable != userID{
		logs.CtxWarn(ctx, "No permission to update this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有更新权限"}
		return resp, nil
	}

	listID := in.ListId
	// 做变更后的唯一性判断
	if in.ListName != nil {
		dupl, musicListID, err := db.IsDuplicateMusicList(ctx, *in.ListName, userID)
		if err != nil && err != gorm.ErrRecordNotFound{ // TODO：这里的判断和处理合适吗？一旦判重失败就不再创建？
			logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
			return resp, err
		}
	
		if dupl && musicListID != listID {
			logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "修改后的歌单名与已有歌单名重复，请修改"}
			return resp, nil
		}
	}

	updateData := map[string]any{}
	utils.AddToMapIfNotNil(updateData, in.ListName) // TODO:入参不对，待修复
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
	resp := &music.GetMusicFromListResponse{}
	// TODO:检查入参的歌单是否存在

	// 鉴权，看请求的歌单是否归属当前登陆人
	userID := utils.GetValue(ctx, "user_id") 
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}

	if userIDFromTable != userID{
		logs.CtxWarn(ctx, "No permission to view this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有查看权限"}
		return resp, nil
	}

	musicList, total, err := db.GetMusicFromList(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	resp.MusicList = musicList // TODO:这里循环赋值吗，有别的处理方案么
	// cannot use musicList (variable of type *[]model.MusicList) as []*music.MusicItem value in assignment
	resp.Total = total
	return resp, nil
} 



func (ml *FeiMusicMusicList)AddMusicToList(ctx context.Context, in *music.AddMusicToListRequest) (*music.AddMusicToListResponse, error){
	// 检查入参中的歌单是否存在，进一步考虑是不是要增加状态字段（软删除），增加的话这里还要判断状态
	resp := &music.AddMusicToListResponse{}
	err := db.JudgeMusicListWithListID(ctx, in.ListId)
	if err != nil{
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil //TODO:这里的错误处理看不出具体的错误是系统错误还是业务错误，也看不出错误的环节，如何改进？
	}

	// TODO：这里是支持批量添加的，目前处理成了单个添加，待修复
	// TODO：不存在的音乐跳过，记录日志，存在的音乐添加，返回中列出添加情况？
	// 检查要添加的音乐是否存在(状态)
	_, err = db.GetMusicWithUniqueMusicID(ctx, in.MusicIds)
	if err != nil{
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil 
	}

	// TODO：加事务？进一步，还有哪里需要加事务吗


	//鉴权，看入参中的歌单是否属于当前登陆人
	userID := utils.GetValue(ctx, "user_id") 
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	}

	if userIDFromTable != userID{
		logs.CtxWarn(ctx, "No permission to add music to this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有添加权限"}
		return resp, nil
	}

	err = db.AddMusicToList(ctx, in.MusicIds, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	} 

	return resp, nil
}

func (ml *FeiMusicMusicList)RemoveMusicFromList(ctx context.Context, in *music.RemoveMusicFromListRequest) (*music.RemoveMusicFromListResponse, error){
	// 检查入参中的歌单是否存在
	resp := &music.RemoveMusicFromListResponse{}
	err := db.JudgeMusicListWithListID(ctx, in.ListId)
	if err != nil{
		logs.CtxWarn(ctx, "failed to delete music from music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
		return resp, nil
	}
	// TODO:支持批量删除、目前做成了仅删除单首，待优化，修复后返回删除情况？

	// 检查要删除的音乐是否存在？有必要吗，应该没必要
	// _, err := db.GetMusicWithUniqueMusicID(ctx, in.MusicID)
	// if err != nil{
	// 	logs.CtxWarn(ctx, "failed to delete music to music_list, list id=%v, music id=%v, err=%v", in.ListID, in.MusicID, err)
	// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
	// 	return resp, nil 
	// }

	// TODO：加事务？进一步，还有哪里需要加事务吗


	//鉴权，看入参中的歌单是否属于当前登陆人
	//实际上，如果没有前序的歌单存在性检查，这里也是可以捕获到这种情况并报错的，是不是前面的检查可以去掉呢
	userID := utils.GetValue(ctx, "user_id") 
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music from music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
		return resp, nil
	}
	if userIDFromTable == "" {
		logs.CtxWarn(ctx, "failed to delete music from music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
		return resp, nil
	}

	if userIDFromTable != userID{
		logs.CtxWarn(ctx, "No permission to delete music from this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有删除权限"}
		return resp, nil
	}

	err = db.DeleteMusicFromList(ctx, in.MusicIds, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music from music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
		return resp, nil
	} 

	return resp, nil
}






