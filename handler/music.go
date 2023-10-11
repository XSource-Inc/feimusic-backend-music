package handler

import (
	"context"

	"github.com/XSource-Inc/feimusic-backend-music/db"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)
// TODO:需要学习如何加监控和追踪（已有，但是处理的感觉不太好）
type FeiMusicMusic struct {
	music.UnimplementedFeiMusicMusicServer
}
// TODO:有些接口限制登陆后访问，有些接口不限制必须登陆才可访问，怎么做呢=>网关层控制的？
func (m *FeiMusicMusic) AddMusic(ctx context.Context, req *music.AddMusicRequest) (*music.AddMusicResponse, error) {
	resp := &user.UserSignUpResponse{}
	musicName := resp.MusicName
	
	err := db.JudgeMusicWithUniqueNameAndArtist(ctx, req.MusicName, req.Artist)
	if err == nil{
		// 发现重名不支持添加
		// logs.CtxWarn(ctx, "the music library already contains this singer's music, singer=%v, music name=%v", req.Artist, req.MusicName)
		// resp.BaseResp =  &base.BaseResp{StatusCode: 1, StatusMessage: "音乐库已包含该歌手的这首音乐"}
		// return resp, nil

		// 或者发现重名后给歌名增加时间戳后缀
		timeStamp := fmt.Sprintf("%d", time.Now().Unix())
		musicName += timeStamp
	}

	if err != nil and err != gorm.ErrRecordNotFound {
		logs.CtxWarn(ctx, "failed to create music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"} 
		return resp, nil 
	}
	
	userID := GetValue(ctx, "user_id")

	newMusic := &model.Music{   
		MusicName: musicName,
		Artist:    req.Artist,
		Album:     req.Album,
		Tags:      req.Tags,
		UserID:    userID,
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

	// 检查要删除的音乐是否存在
	err := JudgeMusicWithMusicID(ctx, req.MusicId)
	if err != nil{
		if err == gorm.ErrRecordNotFound{
			logs.CtxWarn(ctx, "the music to be deleted does not exist, music id=%v", req.MusicId)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要删除的音乐不存在"}
			return resp, nil 
		} else{
			logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
			return resp, nil
		}
	} 

	//TODO:判断是否有删除权限，暂时处理成仅可删除本人上传的音乐


	//TODO:要处理成软删除
	err := db.DeleteMusicWithID(ctx, req.MusicId) // 要删除的记录如果不存在会返回record not found的错误 
	if err != nil{
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil 
	}
	
	return resp, nil
}

func (m *FeiMusicMusic) UpdateMusic(ctx context.Context, req *music.UpdateMusicRequest) (*music.UpdateMusicResponse, error) {
	// TODO:代码结构拆分，目前全写在这一个函数中了
	// 做变更后的唯一性判断
	//TODO:增加权限限制？仅歌曲上传人可修改歌曲？
	// TODO:检查要更新的音乐是否存在
	var (
		change bool
		musicName string = req.MusicName
		artist []string = req.Artist
	)
	
	if musicName != nil && artist != nil {
		change = True
	} else if musicName != nil {
		change = True
		artist = db.GetMusicWithUniqueMusicID(ctx, req.MusicID).Artist
	} else if artist != nil {
		change = True
		musicName = db.GetMusicWithUniqueMusicID(ctx, req.MusicID).MusicName
	} 
	if change {
		err := db.JudgeMusicWithUniqueNameAndArtist(ctx, musicName, artist)
		if err == nil{
			timeStamp := fmt.Sprintf("%d", time.Now().Unix())
			musicName += timeStamp
		}
	}

	resp := &user.MusicUpdateResponse{}

	updateData := map[string]any{}
	utils.AddToMapIfNotNil(updateData, req.MusicName)
	utils.AddToMapIfNotNil(updateData, req.Album)
	utils.AddToMapIfNotNil(updateData, req.Tags)
	utils.AddToMapIfNotNil(updateData, req.Artist)

	err := db.UpdateMusic(ctx, req.MusicID, updateData) // TODO:如果上面if change进入了，这里对err重新声明会有什么问题，有没有其他处理方法
	if err != nil {
		logs.CtxWarn(ctx, "failed to update music, err=%v", err) // TODO:重复报错的问题有没有更好的处理(里层和外层的报错信息可能是不同的？)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
		return resp, nil
	}
	
	return resp, nil
}
// TODO:这个接口还没写, 分页、查询
func (m *FeiMusicMusic) SearchMusic(ctx context.Context, req *music.SearchMusicRequest) (*music.SearchMusicResponse, error) {
	
	resp := &music.SearchMusicResponse{}
	music_list, total, err := db.SearchMusic(ctx, req) // total如果用的是int64，那说明musicid也可以用int64
	if err != nil{
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

	return
}

func (m *FeiMusicMusic) GetMusic(ctx context.Context, req *music.GetMusicRequest) (*music.GetMusicResponse, error) {
	resp := &music.GetMusicResponse{}
	music, err := db.GetMusicWithUniqueMusicID(ctx, music_id)
	if err != nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取音乐失败"}
		return resp, nil
		//TODO:这里是不是应该区分系统错误和要获取的音乐在库中不存在的情况
	}
	if music == nil { //TODO:什么情况下会走到这个分支
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取音乐失败"}
		return resp, nil
	}

	// TODO:不需要返id?
	resp.MusicName = music.MusicName
	resp.Artist = music.Artist
	resp.Album = music.Album
	resp.Tags = music.Tags
	resp.UserID = music.UserId

	resp.URL= temp("根据音乐信息生成音乐路径")// TODO:根据音乐信息生成音乐路径

	return resp, nil
}
