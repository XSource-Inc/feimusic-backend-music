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

type FeiMusicMusic struct {
	music.UnimplementedFeiMusicMusicServer
}

func (m *FeiMusicMusic) AddMusic(ctx context.Context, req *music.AddMusicRequest) (*music.AddMusicResponse, error) {
	resp := &music.AddMusicResponse{}

	// 音乐数据构造
	if req.UserId == 0 {
		logs.CtxWarn(ctx, "failed to add music, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "新增歌曲失败"}
		return resp, nil
	}

	artist := strings.Join(req.Artist, ",")
	if len(artist) > 500 {
		logs.CtxWarn(ctx, "failed to add music, because artist is too log")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "作者名超长，请检查"}
		return resp, nil
	}

	tags := strings.Join(req.Tags, ",")
	if len(tags) > 300 {
		logs.CtxWarn(ctx, "failed to add music, because tags is too log")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "风格超长，请检查"}
		return resp, nil
	}

	musicName := req.MusicName

	newMusic := &model.Music{
		MusicName: musicName,
		Artist:    artist,
		Album:     req.Album,
		Tags:      tags,
		UserID:    req.UserId,
		MD5:       req.Md5,
		Status:    0,
	}

	// 新增音乐
	err := db.AddMusic(ctx, newMusic)
	if err != nil {
		if err == gorm.ErrDuplicatedKey {
			logs.CtxWarn(ctx, "failed to create music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐重复，请确认后重新添加"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to create music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"}
			return resp, nil
		}
	}

	return resp, nil
}

func (m *FeiMusicMusic) MusicDelete(ctx context.Context, req *music.DeleteMusicRequest) (*music.DeleteMusicResponse, error) {
	resp := &music.DeleteMusicResponse{}

	if req.UserId == 0 {
		logs.CtxWarn(ctx, "failed to delete music, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌曲失败"}
		return resp, errors.New("missing user id")
	}

	// 判断是否有删除权限，目前实现：仅支持删除本人上传的音乐
	music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil
	}
	if music.UserID != req.UserId {
		logs.CtxWarn(ctx, "failed to delete music, no authority to delete music from others")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil
	}

	//开启事物，软删除音乐和歌单音乐
	err = db.GetDB().Transaction(func(tx *gorm.DB) error {

		err = db.DeleteMusicWithID(ctx, tx, req.MusicId, req.UserId)
		if err != nil {
			logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
			return err
		}

		err = db.DeleteMusicFromList(ctx, tx, req.MusicId)
		if err != nil {
			logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
			return err
		}

		return nil
	})

	if err != nil {
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil
	}

	return resp, nil
}

func (m *FeiMusicMusic) UpdateMusic(ctx context.Context, req *music.UpdateMusicRequest) (*music.UpdateMusicResponse, error) {
	resp := &music.UpdateMusicResponse{}

	if req.UserId == 0 {
		logs.CtxWarn(ctx, "failed to update music, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌曲失败"}
		return resp, errors.New("missing user id")
	}

	// 判断是否有更新权限，目前实现：仅支持更新本人上传的音乐
	music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
		return resp, nil
	}
	if music.UserID != req.UserId {
		logs.CtxWarn(ctx, "failed to update music, no authority to update music from others")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
		return resp, nil
	}

	// 构建数据
	updateData := map[string]any{}
	tags := strings.Join(req.Tags, ",")
	updateData["tags"] = tags
	artist := strings.Join(req.Artist, ",")
	updateData["artist"] = artist
	utils.AddToMapIfNotNil(updateData, req.MusicName, "music_name")
	utils.AddToMapIfNotNil(updateData, req.Album, "album")

	err = db.UpdateMusic(ctx, req.MusicId, updateData)
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
		return resp, nil
	}

	return resp, nil
}

func (m *FeiMusicMusic) SearchMusic(ctx context.Context, req *music.SearchMusicRequest) (*music.SearchMusicResponse, error) {
	resp := &music.SearchMusicResponse{}

	musicList, total, err := db.SearchMusic(ctx, req)
	if err != nil {
		logs.CtxWarn(ctx, "failed to search music")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "搜索音乐失败"}
		return resp, nil 
	}

	musicItems := []*music.MusicItem{}
	for _, m := range *musicList {
		musicItem := music.MusicItem{}
		musicItem.MusicId = m.MusicID
		musicItem.MusicName = m.MusicName
		musicItem.Artist = strings.Split(m.Artist, ",")
		musicItem.Album = m.Album
		musicItem.Tags = strings.Split(m.Tags, ",")
		musicItem.UserId = m.UserID
		musicItems = append(musicItems, &musicItem)
	}

	resp.MusicList = musicItems
	resp.Total = total

	return resp, nil
}


