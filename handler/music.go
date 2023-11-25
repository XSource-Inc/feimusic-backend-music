package handler

import (
	"context"
	"errors"
	"sort"
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

// 使用音乐名和歌手联合判重
// TODO：修改方案。使用歌手名和音乐名做联合主键，新增时如果重复会报错，不对音乐做重复处理（记录MD5还有必要吗）
// 2、幂等处理：请求参数增加字节流（音乐文件），先校验md5是否重复，重复不支持写入，直接报错，不重复就入库

// err := db.JudgeMusicWithUniqueNameAndArtist(ctx, musicName, artist)
// if err == nil {
// 	// 发现重名不支持添加
// 	logs.CtxWarn(ctx, "the music library already contains this singer's music, singer=%v, music name=%v", req.Artist, req.MusicName)
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐重复，请修改音乐名后重新添加"}
// 	return resp, nil
// }

// if err != nil && err != gorm.ErrRecordNotFound {
// 	logs.CtxWarn(ctx, "failed to create music, err=%v", err)
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"}
// 	return resp, nil
// }

func (m *FeiMusicMusic) AddMusic(ctx context.Context, req *music.AddMusicRequest) (*music.AddMusicResponse, error) {
	resp := &music.AddMusicResponse{}

	// 音乐数据构造
	if req.UserId == 0 {
		logs.CtxWarn(ctx, "failed to add music, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "新增歌曲失败"}
		return resp, nil
	}

	sort.Strings(req.Artist)
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
		Status:    0,
	}

	// 新增音乐
	err := db.AddMusic(ctx, newMusic)
	if err != nil { // TODO：待升级，使用postgresql数据库，这个数据库咋整呀，我现在还没有数据库呢，都不知道咋整， 搜gorm postgres就可以了
		if err == gorm.ErrDuplicatedKey { // 代码里面没有需要改造的东西是吧
			logs.CtxWarn(ctx, "failed to create music, err=%v", err) // w我很想建个表，把数据库搭好，但是我现在表结构还不知道合不合适
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐重复，请确认后重新添加"}
			return resp, nil
		} else {
			logs.CtxWarn(ctx, "failed to create music, err=%v", err)
			resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "添加音乐失败"}
			return resp, nil
		}
	}
	// TODO：新增音乐字节流字段，并把音乐文件写入硬盘，1、音乐咋存，直接转成二进制存在music中（判断大小，太大先报错不存）2、md5还需要吗（目前不知道哪里用的到）

	return resp, nil
}

// 冗余的逻辑
// music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId) // TODO：不单独写个函数了吧
// userIDOfMusic := music.UserID
// if err != nil {
// 	// 检查要删除的音乐是否存在
// 	if err == gorm.ErrRecordNotFound {
// 		logs.CtxWarn(ctx, "the music to be deleted does not exist, music id=%v", req.MusicId)
// 		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要删除的音乐不存在"}
// 		return resp, nil
// 	} else {
// 		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
// 		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
// 		return resp, nil
// 	}
// }
// if userID != userIDOfMusic {
// 	logs.CtxWarn(ctx, "currently, deleting music uploaded by others is not supported, music id=%v, operator id=%v", req.MusicId, userID)
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "暂不支持删除他人上传的音乐"}
// 	return resp, nil
// }

// TODO：遗漏一个重要逻辑，删除音乐的同时，怎么处理歌单中的音乐呢=》设置外键后，修改音乐状态，音乐列表和音乐的关联表的状态会自动修改？
// 怎么设置这个外键，另外，好像看到过一个说法是尽量不要使用gorm的外键约束？
// TODO：目前没支持做批量删除，后续需要时再添加
func (m *FeiMusicMusic) MusicDelete(ctx context.Context, req *music.DeleteMusicRequest) (*music.DeleteMusicResponse, error) {
	resp := &music.DeleteMusicResponse{}

	//判断是否有删除权限，目前处理成仅可删除本人上传的音乐
	if req.UserId == 0 {
		logs.CtxWarn(ctx, "failed to delete music, because the user id was not obtained")
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除歌曲失败"}
		return resp, errors.New("missing user id")
	}

	//处理成软删除
	err := db.DeleteMusicWithID(ctx, req.MusicId, req.UserId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"}
		return resp, nil
	}

	// TODO：先直接处理，外键那个后边看
	err = db.DeleteMusicFromList(ctx, req.MusicId)
	if err != nil {
		logs.CtxWarn(ctx, "failed to delete music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "删除音乐失败"} // TODO:这里的处理合适吗，前序删除已经成功了=》加事务？
		return resp, nil
	}
	// 这里加事务的话怎么组织代码结构呢？
	// 不太会写事务，参照下面的代码，咋写呢，我咋没有tx *gorm.DB参数呢？
	// db.Transaction(func(tx *gorm.DB) error {
	// 	// 在事务中执行更新操作
	// 	if err := tx.Model(&Music{}).Where("music_id = ?", musicID).Update("status", newStatus).Error; err != nil {
	// 		// 更新失败，回滚事务
	// 		return err
	// 	}
	// 	// 更新成功，事务会自动提交
	// 	return nil
	// })

	return resp, nil
}

//权限限制：先不做本人权限限制了
// userID := req.UserId
// if len(userID) == 0 {
// 	logs.CtxWarn(ctx, "failed to update music, because the user id was not obtained")
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新歌曲失败"}
// 	return resp, errors.New("missing user id")
// }
// music, err := db.GetMusicWithUniqueMusicID(ctx, req.MusicId)
// if err != nil {
// 	logs.CtxWarn(ctx, "failed to update music, err=%v", err)
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
// 	return resp, nil
// }
// userIDOfMusic := music.UserID
// if err != nil {
// 	// 检查要更新的音乐是否存在
// 	if err == gorm.ErrRecordNotFound {
// 		logs.CtxWarn(ctx, "the music to be updated does not exist, music id=%v", req.MusicId)
// 		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "要更新的音乐不存在"}
// 		return resp, nil
// 	} else {
// 		logs.CtxWarn(ctx, "failed to update music, err=%v", err)
// 		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "更新音乐失败"}
// 		return resp, nil
// 	}
// }
// if userID != userIDOfMusic {
// 	logs.CtxWarn(ctx, "currently, update music uploaded by others is not supported, music id=%v, operator id=%v", req.MusicId, userID)
// 	resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "暂不支持更新他人上传的音乐"}
// 	return resp, nil
// }

// 做变更后的唯一性判断
// TODO:也可以把这部分工作交给数据库做
// var (
// 	change    bool
// 	musicName string   = *req.MusicName
// 	artist    []string = req.Artist
// )

// if musicName != "" && artist != nil {
// 	change = true
// } else if musicName != "" {
// 	change = true
// 	artist = music.Artist
// } else if artist != nil {
// 	change = true
// 	musicName = music.MusicName
// }
// if change {
// 	err = db.JudgeMusicWithUniqueNameAndArtist(ctx, musicName, artist)
// 	if err == nil {
// 		timeStamp := fmt.Sprintf("%d", time.Now().Unix())
// 		musicName += timeStamp
// 		// TODO：待修改，直接报错
// 	}

// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		logs.CtxWarn(ctx, "failed to update music, err=%v", err)
// 		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "音乐更新失败"}
// 		return resp, nil
// 	}
// }

func (m *FeiMusicMusic) UpdateMusic(ctx context.Context, req *music.UpdateMusicRequest) (*music.UpdateMusicResponse, error) {
	// TODO:代码结构拆分，目前全写在这一个函数中了
	resp := &music.UpdateMusicResponse{}

	updateData := map[string]any{}
	tags := strings.Join(req.Tags, ",")
	updateData["tags"] = tags
	artist := strings.Join(req.Artist, ",")
	updateData["artist"] = artist
	utils.AddToMapIfNotNil(updateData, req.MusicName, "music_name")
	utils.AddToMapIfNotNil(updateData, req.Album, "album")

	err := db.UpdateMusic(ctx, req.MusicId, updateData)
	if err != nil {
		// TODO:这里应该做唯一索引冲突的判断，但是目前找不到对应的错误，应该是gorm版本的问题
		// if errors.Is(err, gorm.er)
		// 多个位置调用update music，下面这行日志可以用来区分调用的位置
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
		return resp, nil // TODO:不太对劲，error全返回nil了，这个变量还有啥意义
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

// TODO:还差个音乐资源
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

	if music == nil {
		logs.CtxWarn(ctx, "failed to get music, err=%v", err)
		resp.BaseResp = &base.BaseResp{StatusCode: 1, StatusMessage: "获取音乐失败"}
		return resp, nil
	}

	artist := strings.Split(music.Artist, ",")
	tags := strings.Split(music.Tags, ",")
	resp.MusicId = music.MusicID
	resp.MusicName = music.MusicName
	resp.Artist = artist
	resp.Album = music.Album
	resp.Tags = tags
	resp.UserId = music.UserID

	// resp.Url = temp("根据音乐信息生成音乐路径") // TODO:根据音乐信息生成音乐路径。存储路径+MD5

	return resp, nil
}
