package main

import (
	"net"

	"google.golang.org/grpc"
	"github.com/XSource-Inc/grpc_idl/go/proto_gen/fei_music/music"
)

func main() {
	lis, err := net.Listen("tcp", ":8082")
	if err != nil{
		panic(err)
	}
	srv := grpc.NewServer()
	music.RegisterFeiMusicMusicServer(srv, &FeiMusicMusic{})
	if err := srv.Serve(lis); err != nil{
		panic(err)
	}
}