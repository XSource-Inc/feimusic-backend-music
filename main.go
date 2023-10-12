package main

import (
	"net"

	"google.golang.org/grpc"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
	"github.com/XSource-Inc/feimusic-backend-music/handler"
)

func main() {
	lis, err := net.Listen("tcp", ":8082")
	if err != nil{
		panic(err)
	}
	srv := grpc.NewServer()
	music.RegisterFeiMusicMusicServer(srv, &handler.FeiMusicMusic{}) //TODO：音乐和音乐列表是要注册两个服务？
	if err := srv.Serve(lis); err != nil{
		panic(err)
	}
}