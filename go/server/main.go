package main

import (
	"fmt"
	"net"
	"os"

	// protoc で自動生成されたパッケージ
	"github.com/yamagai/grpc-tutorial/go/deepthought"
	"google.golang.org/grpc"
)

const portNumber = 13333

func main() {
	serv := grpc.NewServer()

	// 実装した Server を登録
	deepthought.RegisterComputeServer(serv, &Server{})

	// 待ち受けソケットを作成
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", portNumber))
	if err != nil {
		fmt.Println("failed to listen:", err)
		os.Exit(1)
	}

	// gRPC サーバーでリクエストの受付を開始
	// l は Close されてから戻るので、main 関数での Close は不要
	serv.Serve(l)
}
