package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/Kidsunbo/kie_toolbox_go/logs"
	"github.com/XSource-Inc/feimusic-backend-music/db"
	"github.com/XSource-Inc/feimusic-backend-music/model"
	"github.com/XSource-Inc/feimusic-backend-music/utils"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/base"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
	"github.com/jinzhu/gorm"
)

// TODO:需要学习如何加监控和追踪（已有，但是处理的感觉不太好）
// TODO:参数的处理，例如去重前后空格，放在哪里处理呢？
type FeiMusicMusic struct {
	music.UnimplementedFeiMusicMusicServer
}

// TODO:有些接口限制登陆后访问，有些接口不限制必须登陆才可访问，怎么做呢=>网关层控制的？
func (m *FeiMusicMusic) AddMusic(ctx context.Context, req *music.AddMusicRequest) (*music.AddMusicResponse, error) {
	resp := &music.AddMusicResponse{}
	musicName := req.MusicName
	// 使用音乐名和歌手联合判重
	err := db.JudgeMusicWithUniqueNameAndArtist(ctx, req.MusicName, req.Artist)
	if err == nil {
		// 发现重名不支持添加
		// logs.CtxWarn(ctx, "the music library already contains this singer's music, singer=%v, music name=%v", req.Artist, req.MusicName)
		// resp.BaseResp =  &base.BaseResp{StatusCode: 1, StatusMessage: "音乐库已包含该歌手的这首音乐"}
		// return resp, nil

		// 或者发现重名后给歌名增加时间戳后缀
		timeStamp := fmt.Sprintf("%d", time.Now().Unix())
		musicName += timeStamp
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		logs.CtxWarn(ctx, "failed to create music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"}
		return resp, nil
	}

	userID := utils.GetValue(ctx, "user_id")

	newMusic := &model.Music{
		MusicName: musicName,
		Artist:    req.Artist,
		Album:     req.Album,
		Tags:      req.Tags,
		UserID:    userID,
		Status:    0,
	}
	err = db.AddMusic(ctx, newMusic)
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"}
		return resp, nil
	}

	return resp, nil
}

func (m *FeiMusicMusic) MusicDelete(ctx context.Context, req *music.DeleteMusicRequest) (*music.DeleteMusicResponse, error) {
	resp := &music.DeleteMusicResponse{}

	//判断是否有删除权限，暂时处理成仅可删除本人上传的音乐
	userID := utils.GetValue(ctx, "user_id")
	music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId) // TODO：不单独写个函数了吧
	userIDOfMusic := music.UserID
	if err != nil {
		// 检查要删除的音乐是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "the music to be deleted does not exist, music id=%v", req.MusicId)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要删除的音乐不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
			return resp, nil
		}
	}
	if userID != userIDOfMusic {
		logs.CtxWarn(ctx, "currently, deleting music uploaded by others is not supported, music id=%v, operator id=%v", req.MusicId, userID)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "暂不支持删除他人上传的音乐"}
		return resp, nil
	}

	//处理成软删除
	err = db.DeleteMusicWithID(ctx, req.MusicId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil
	}

	return resp, nil
}

func (m *FeiMusicMusic) UpdateMusic(ctx context.Context, req *music.UpdateMusicRequest) (*music.UpdateMusicResponse, error) {
	// TODO:代码结构拆分，目前全写在这一个函数中了
	resp := &music.UpdateMusicResponse{}
	//权限限制：暂定仅歌曲上传人可修改歌曲
	userID := utils.GetValue(ctx, "user_id")
	music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId) 
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
		return resp, nil
	}
	userIDOfMusic := music.UserID
	if err != nil {
		// 检查要更新的音乐是否存在
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "the music to be updated does not exist, music id=%v", req.MusicId)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要更新的音乐不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to update music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
			return resp, nil
		}
	}
	if userID != userIDOfMusic {
		logs.CtxWarn(ctx, "currently, update music uploaded by others is not supported, music id=%v, operator id=%v", req.MusicId, userID)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "暂不支持更新他人上传的音乐"}
		return resp, nil
	}

	// 做变更后的唯一性判断
	var (
		change    bool
		musicName string   = *req.MusicName
		artist    []string = req.Artist
	)

	if musicName != "" && artist != nil {
		change = true
	} else if musicName != "" {
		change = true
		artist = music.Artist
	} else if artist != nil {
		change = true
		musicName = music.MusicName
		if err != nil {
			logs.CtxWarn(ctx, "failed to update music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
			return resp, nil
		}
	}
	if change {
		err = db.JudgeMusicWithUniqueNameAndArtist(ctx, musicName, artist)
		if err == nil {
			timeStamp := fmt.Sprintf("%d", time.Now().Unix())
			musicName += timeStamp
		}

		if err != nil && err != gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to update music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
			return resp, nil
		}
	}

	updateData := map[string]any{}
	utils.AddToMapIfNotNil(updateData, req.MusicName) // TODO:入参不对，待修复
	utils.AddToMapIfNotNil(updateData, req.Album)
	utils.AddToMapIfNotNil(updateData, req.Tags)
	utils.AddToMapIfNotNil(updateData, req.Artist)

	err = db.UpdateMusic(ctx, req.MusicId, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", err) // TODO:重复报错的问题有没有更好的处理(里层和外层的报错信息可能是不同的？)
		// TODO：有必要区分报错的位置吗，日志中本身就会标记文件、函数名和代码行？
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
		return resp, nil
	}

	return resp, nil
}

// TODO:这个接口还没写, 分页、查询
func (m *FeiMusicMusic) SearchMusic(ctx context.Context, req *music.SearchMusicRequest) (*music.SearchMusicResponse, error) {

	resp := &music.SearchMusicResponse{}
	music_list, total, err := db.SearchMusic(ctx, req) // total如果用的是int64，那说明musicid也可以用int64
	if err != nil {
		logs.CtxWarn(ctx, "failed to search music")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return nil, err
	}
	// MusicList []*MusicItem  不会写了 music_list怎么转成MusicList?for 循环把model.music转成MusicItem?

	// musicItem中有musicID!!!!!!
	// type MusicItem struct {
	// 	state         protoimpl.MessageState
	// 	sizeCache     protoimpl.SizeCache
	// 	unknownFields protoimpl.UnknownFields

	// 	MusicId   string   `protobuf:"bytes,1,opt,name=music_id,json=musicId,proto3" json:"music_id,omitempty"`
	// 	MusicName string   `protobuf:"bytes,2,opt,name=music_name,json=musicName,proto3" json:"music_name,omitempty"`
	// 	Artist    []string `protobuf:"bytes,3,rep,name=artist,proto3" json:"artist,omitempty"`
	// 	Album     string   `protobuf:"bytes,4,opt,name=album,proto3" json:"album,omitempty"`
	// 	Tags      []string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
	// 	UserId    string   `protobuf:"bytes,6,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	// }

	return nil, nil
}

func (m *FeiMusicMusic) GetMusic(ctx context.Context, req *music.GetMusicRequest) (*music.GetMusicResponse, error) {
	resp := &music.GetMusicResponse{}
	music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.CtxWarn(ctx, "failed to get music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "指定获取的音乐不存在"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to get music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取音乐失败"}
			return resp, nil
		}
	}

	// 防御式编程，避免gorm不符预期
	// TODO:检查一下是不是还有别的地方要补充防御式代码
	if music == nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取音乐失败"}
		return resp, nil
	}

	resp.MusicId = music.MusicID
	resp.MusicName = music.MusicName
	resp.Artist = music.Artist
	resp.Album = *music.Album
	resp.Tags = music.Tags
	resp.UserId = music.UserID

	resp.Url = temp("根据音乐信息生成音乐路径") // TODO:根据音乐信息生成音乐路径

	return resp, nil
}
