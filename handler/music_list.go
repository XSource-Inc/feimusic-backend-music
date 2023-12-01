package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/db"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/feimusic-backend-music/utils"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/base"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
	"gorm.io/gorm"
)

type FeiMusicMusicList struct {
	music.UnimplementedFeiMusicMusicServer
}
const (
	success = 1
)

// TODO:重要逻辑缺陷，创建歌单时，应先检查对应的歌单是否在表中存在，存在直接修改状态，不存在再新建。或者有别的更合适的方案吗？
func (ml *FeiMusicMusicList) CreateMusicList(ctx context.Context, in *music.CreateMusicListRequest) (*music.CreateMusicListResponse, error) {
	resp := &music.CreateMusicListResponse{}

	// 参数校验
	// TODO: user_id有效性的校验
	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to create music list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
		return resp, nil
	}
	
	if len(in.ListName) > 500 {
		logs.CtxWarn(ctx, "failed to add music list, because music list name is too log")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单名超长，请检查"}
		return resp, nil
	}

	if len(*in.ListComment) > 1000 {
		logs.CtxWarn(ctx, "failed to add music list, because music list comment is too log")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单简介超长，请检查"}
		return resp, nil
	}

	tags := strings.Join(in.Tags, ",")
	if len(tags) > 300 {
		logs.CtxWarn(ctx, "failed to add music list, because tags is too log")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单风格超长，请检查"}
		return resp, nil
	}

	// 添加歌单记录
	newMusicList := &model.MusicList{
		ListName:    in.ListName,
		ListComment: in.ListComment,
		Tags:        tags,
		UserID:      in.UserId,
		Status:      success,
	}

	listID, err := db.CreateMusicList(ctx, newMusicList)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单重复，请确认后重新添加"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "创建歌单失败"}
			return resp, nil
		}
	}

	resp.ListId = listID
	return resp, nil
}

func (ml *FeiMusicMusicList) DeleteMusicList(ctx context.Context, in *music.DeleteMusicListRequest) (*music.DeleteMusicListResponse, error) {
	// 删除歌单时仅限制操作人是歌单归属人
	resp := &music.DeleteMusicListResponse{}

	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to delete music list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
		return resp, nil
	}

	err := db.GetDB().Transaction(func(tx *gorm.DB) error {

		// 软删除,指定歌单id和归属用户id
		err := db.DeleteMusicList(ctx, tx, in.ListId, in.UserId)
		if err != nil {
			logs.CtxWarn(ctx, "failed to delete music list, list id=%v, user id=%v, err=%v", in.ListId, in.UserId, err)
			return err
		}

		// 软删除-删除歌单下歌曲
		err = db.DeleteListMusic(ctx, tx, in.ListId) 
		if err != nil {
			logs.CtxWarn(ctx, "failed to delete songs from list, listid=%v, err=%v", in.ListId, err)
			return err
		}
		return nil
	})
	// TODO:使用事务之后，对于错误信息的把控精度变低了
	if err != nil{
		if err == gorm.ErrRecordNotFound {
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要删除的歌单不存在，请确认"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to create music list, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单失败"}
			return resp, nil
		}
	}
	
	return resp, nil
}

func (ml *FeiMusicMusicList) UpdateMusicList(ctx context.Context, in *music.UpdateMusicListRequest) (*music.UpdateMusicListResponse, error) {
	resp := &music.UpdateMusicListResponse{BaseResp: &base.BaseResp{}}

	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to update music list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
		return resp, errors.New("missing user id")
	}

	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to update music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要更新的歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to update music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
			return resp, nil
		}
	}

	if userIDFromTable != in.UserId {
		logs.CtxWarn(ctx, "the music_list does not belong to the operator, No permission to update")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有更新权限"}
		return resp, nil
	}

	listID := in.ListId
	// 做变更后的唯一性判断 ==》TODO:把这部分工作改到数据库做
	if in.ListName != nil {
		dupl, musicListID, err := db.IsDuplicateMusicList(ctx, *in.ListName, in.UserId)
		if err != nil {
			logs.CtxWarn(ctx, "failed to update music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌单失败"}
			return resp, err
		}

		if dupl && musicListID != listID {
			logs.CtxWarn(ctx, "failed to update music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "修改后的歌单名与已有歌单名重复，请修改"}
			return resp, nil
		}
	}

	tags := strings.Join(in.Tags, ",")

	updateData := map[string]any{}
	updateData["tags"] = tags
	utils.AddToMapIfNotNil(updateData, in.ListName, "listName")
	utils.AddToMapIfNotNil(updateData, in.ListComment, "listComment")

	err = db.UpdateMusicList(ctx, listID, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单信息更新失败"}
		return resp, nil
	}

	return resp, nil
}

// TODO:目前没有返回歌单的基础信息,要做成两个接口还是一个接口呢？
func (ml *FeiMusicMusicList) GetMusicFromList(ctx context.Context, in *music.GetMusicFromListRequest) (*music.GetMusicFromListResponse, error) {
	resp := &music.GetMusicFromListResponse{}

	// 鉴权，看请求的歌单是否归属当前登陆人
	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to get music from the list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, errors.New("missing user id")
	}

	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		// 检查歌单是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to get music from music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to get music from music list, list id=%v, err=%v", in.ListId, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
			return resp, nil
		}
	}

	if userIDFromTable != in.UserId {
		logs.CtxWarn(ctx, "the music_list does not belong to the operator, No permission to view")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有查看权限"}
		return resp, nil
	}

	// 获取歌单信息

	// 获取音乐信息
	musicIDs, err := db.GetMusicFromList(ctx, in.ListId, 0) //TODO:创建常量
	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, list id=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	resp.Total = int64(len(musicIDs))
	if resp.Total == 0 {
		return resp, nil
	}

	musicList, err := db.BatchGetMusicWithMsuicID(ctx, musicIDs)
	resp.MusicList = musicList

	if err != nil {
		logs.CtxWarn(ctx, "failed to obtain music of music_list, listid=%v, err=%v", in.ListId, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取歌单音乐失败"}
		return resp, nil
	}
	if len(resp.MusicList) == 0 {
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

	// 鉴权，看请求的歌单是否归属当前登陆人
	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to add music to the list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加歌曲失败"}
		return resp, errors.New("missing user id")
	}

	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		// 检查入参中的歌单是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to add music to the list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to add music to the list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
			return resp, nil
		}
	}

	if userIDFromTable != in.UserId {
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
		// 如果指定添加的字段中存在无效的音乐id，目前只在日志中做了记录
		logs.CtxWarn(ctx, "there is invalid music in the specified added music, music id=%v", invalidMusicIDs)
	}

	err = db.AddMusicToList(ctx, in.ListId, effectiveMusicIDs)
	if err != nil {
		logs.CtxWarn(ctx, "failed to add music to music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "向歌单中添加音乐失败"}
		return resp, nil
	}

	return resp, nil
}

func (ml *FeiMusicMusicList) RemoveMusicFromList(ctx context.Context, in *music.RemoveMusicFromListRequest) (*music.RemoveMusicFromListResponse, error) {
	resp := &music.RemoveMusicFromListResponse{}

	// 鉴权，看请求的歌单是否归属当前登陆人
	if in.UserId == 0 {
		logs.CtxWarn(ctx, "failed to delete music form list, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌曲失败"}
		return resp, errors.New("missing user id")
	}

	userIDFromTable, err := db.GetUserIDWithListID(ctx, in.ListId)
	if err != nil {
		// 检查入参中的歌单是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to delete music form the music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "歌单不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to delete music form the music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "从歌单中删除音乐失败"}
			return resp, nil
		}
	}

	if userIDFromTable != in.UserId {
		logs.CtxWarn(ctx, "No permission to delete music from the music_list, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "非本人歌单，没有删除权限"}
		return resp, nil
	}

	// 软删除
	err = db.BatchUpdateMusicStatus(ctx, in.ListId, in.MusicIds, 1)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music from the music_list, list id=%v, music id=%v, err=%v", in.ListId, in.MusicIds, err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌单中指定音乐失败"}
		return resp, nil
	}

	return resp, nil
}
