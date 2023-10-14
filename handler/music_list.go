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

}

func (ml *FeiMusicMusicList) CreateMusicList(ctx context.Context, in *music.CreateMusicListRequest) (*music.CreateMusicListResponse, error) {
	userID := utils.GetValue(ctx, "user_id")

	resp := &music.CreateMusicListResponse{}
	// TODO:代码结构调整
	dupl, _, err := db.IsDuplicateMusicList(ctx, in.ListName, userID)
	if err != nil { 
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
		ListName:    in.ListName,
		MusicIDs:    []string{},
		ListComment: in.ListComment,
		Tags:        in.Tags,
		UserID:      userID,
	}

	listID, err := db.CreateMusicList(ctx, newMusicList)
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, nil
	}
	resp.ListId = listID
	return resp, nil
}

func (ml *FeiMusicMusicList) DeleteMusicList(ctx context.Context, in *music.DeleteMusicListRequest) (*music.DeleteMusicListResponse, error) {
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

	if userIDFromTable != userID {
		logs.CtxWarn(ctx, "No permission to delete this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有删除权限"}
		return resp, nil
	}
	// 软删除
	err = db.DeleteMusicList(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}
	return resp, nil
}

func (ml *FeiMusicMusicList) UpdateMusicList(ctx context.Context, in *music.UpdateMusicListRequest) (*music.UpdateMusicListResponse, error) {
	resp := &music.UpdateMusicListResponse{}
	
	// 仅可更新本人歌单  
	userID := utils.GetValue(ctx, "user_id") // TODO:需要考虑取不到userid的情况吗=》需要
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要更新的歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to update music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
			return resp, nil
		}
	}

	if userIDFromTable != userID {
		logs.CtxWarn(ctx, "No permission to update this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有更新权限"}
		return resp, nil
	}

	listID := in.ListId
	// 做变更后的唯一性判断
	if in.ListName != nil {
		dupl, musicListID, err := db.IsDuplicateMusicList(ctx, *in.ListName, userID)
		if err != nil {
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
	utils.AddToMapIfNotNil(updateData, in.ListName, "listName")
	utils.AddToMapIfNotNil(updateData, in.ListComment, "listComment")
	utils.AddToMapIfNotNil(updateData, in.Tags, "tags") // TODO:错误，待修复

	err = db.UpdateMusicList(ctx, listID, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单信息更新失败"}
		return resp, nil
	}
	// TODO：需要返listid吗？
	return resp, nil
}
// TODO：没有处理已删除的音乐
func (ml *FeiMusicMusicList) GetMusicFromList(ctx context.Context, in *music.GetMusicFromListRequest) (*music.GetMusicFromListResponse, error) {
	resp := &music.GetMusicFromListResponse{}

	// 鉴权，看请求的歌单是否归属当前登陆人
	userID := utils.GetValue(ctx, "user_id")
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		// 检查歌单是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to get music from music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to get music from music list, listid=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
			return resp, nil
		}
	}

	if userIDFromTable != userID {
		logs.CtxWarn(ctx, "No permission to view this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有查看权限"}
		return resp, nil
	}

	musicIDs, err := db.GetMusicFromList(ctx, in.ListId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	resp.Total = int64(len(musicIDs))
	if resp.Total == 0 {
		return resp, nil
	}
	
	musicList, err := db.BatchGetMusicWithMsuicID(ctx, musicIDs)
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	if len(resp.MusicList) == 0{
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	resp.MusicList = musicList

	return resp, nil
}

func (ml *FeiMusicMusicList) AddMusicToList(ctx context.Context, in *music.AddMusicToListRequest) (*music.AddMusicToListResponse, error) {
	resp := &music.AddMusicToListResponse{}

	// TODO：加事务？进一步，还有哪里需要加事务吗

	//鉴权，看入参中的歌单是否属于当前登陆人
	userID := utils.GetValue(ctx, "user_id")
	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		// 检查入参中的歌单是否存在
		if err == gorm.ErrRecordNotFound{
			logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
			return resp, nil
		}
	}

	if userIDFromTable != userID {
		logs.CtxWarn(ctx, "No permission to add music to this music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有添加权限"}
		return resp, nil
	}

	// 筛选出入参中有效状态的音乐
	effectiveMusicIDs, invalidMusicIDs, err := db.FilterMusicIDUsingIDAndStatus(ctx, in.MusicIds)

	if err != nil {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	}

	if len(effectiveMusicIDs) == 0 {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "指定添加的音乐均不存在"}
		return resp, nil
	}

	if len(invalidMusicIDs) > 0 {
		// TODO:如果指定添加的字段中存在无效的音乐id，需要返回吗，目前只在日志中做了记录
		logs.CtxWarn(ctx, "there is invalid music in the specified added music, music id=%v", invalidMusicIDs)
	}

	err = db.AddMusicToList(ctx, in.ListId, in.MusicIds)
	if err != nil {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	}

	return resp, nil
}

func (ml *FeiMusicMusicList) RemoveMusicFromList(ctx context.Context, in *music.RemoveMusicFromListRequest) (*music.RemoveMusicFromListResponse, error) {
	// 检查入参中的歌单是否存在
	resp := &music.RemoveMusicFromListResponse{}
	err := db.JudgeMusicListWithListID(ctx, in.ListId)
	if err != nil {
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

	if userIDFromTable != userID {
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
