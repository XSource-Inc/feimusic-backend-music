package handler

import (
	"context"

	"github.com/XSource-Inc/feimusic-backend-music/db"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)

type FeiMusicMusic struct {
	music.UnimplementedFeiMusicMusicServer
}

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
	
	userID := GetValue(ctx, "user_id")

	newMusic := &model.Music{   
		MusicName: musicName,
		Artist:    req.Artist,
		Album:     req.Album,
		Tags:      req.Tags,
		UserID:    userID,
	}
	err = db.AddMusic(ctx, newMusic) // TODO:这里为什么没有正常跳转
	if err != nil {
		logs.CtxWarn(ctx, "failed to create music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"} 
		return resp, nil // TODO:这里其实重复了，删除吗？
	} 

	return resp, nil 
}

// TODO:音乐记录要做假删除吗
func (m *FeiMusicMusic) MusicDelete(ctx context.Context, req *music.DeleteMusicRequest) (*music.DeleteMusicResponse, error) {
	resp := &music.DeleteMusicResponse{}
	err := db.DeleteMusicWithID(ctx, req.MusicId) // 要删除的记录如果不存在会返回record not found的错误
	if err != nil{
		logs.CtxWarn(ctx, "failed to deletd music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
	}
	
	return resp, nil
}

func (m *FeiMusicMusic) UpdateMusic(ctx context.Context, req *music.UpdateMusicRequest) (*music.UpdateMusicResponse, error) {
	// TODO:代码结构拆分，目前全写在这一个函数中了
	// 做变更后的唯一性判断
	var (
		change bool
		musicName string = req.MusicName
		artist []string = req.Artist
	)
	
	if musicName != nil and artist != nil {
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
		logs.CtxWarn(ctx, "failed to update music, err=%v", err) // TODO:重复报错的问题有没有更好的处理
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
		return resp, nil
	}
	
	return resp, nil
}

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
		logs.CtxWarn(ctx, "")// 这么写都不对呀，上面的函数里面已经打过一个日志了，这里咋还打日志呢
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, err
	}
	if music == nil {
		logs.CtxWarn(ctx, "")//咋写？
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "?"}
		return resp, error.New("查询的音乐不存在，请确认")
	}

	// 不需要返id?
	resp.MusicName = music.MusicName,
	resp.Artist = music.Artist,
	resp.Album = music.Album,
	resp.Tags = music.Tags,
	resp.UserID = music.UserId,

	resp.URL= temp("根据音乐信息生成音乐路径")

	return resp, nil
}
