package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yamagai/grpc-tutorial/go/deepthought"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	err := subMain()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func subMain() error {
	if len(os.Args) != 2 {
		return errors.New("usage: client HOST:PORT")
	}
	// コマンドライン引数で渡されたアドレスに接続
	addr := os.Args[1]

	// grpc.WithInsecure() を指定することで、TLS ではなく平文で接続
	// 通信内容が保護できないし、不正なサーバーに接続しても検出できないので本当はダメ
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	// 使い終わったら Close しないとコネクションがリークします
	defer conn.Close()

	// 自動生成された RPC クライアントを conn から作成
	// gRPC は HTTP/2 の stream を用いるため、複数のクライアントが同一の conn を使えます。
	// また RPC クライアントのメソッドも複数同時に呼び出し可能です。
	// see https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md
	cc := deepthought.NewComputeClient(conn)

	// Boot を 2.5 秒後にクライアントからキャンセルするコード
	ctx, cancel := context.WithCancel(context.Background())
	go func(cancel func()) {
		time.Sleep(2500 * time.Millisecond)
		cancel()
	}(cancel)

	// 自動生成された Boot RPC 呼び出しコードを実行
	stream, err := cc.Boot(ctx, &deepthought.BootRequest{})
	if err != nil {
		return err
	}

	// ストリームから読み続ける
	for {
		resp, err := stream.Recv()
		if err != nil {
			// io.EOF は stream の正常終了を示す値
			if err == io.EOF {
				break
			}
			// status パッケージは error と gRPC status の相互変換を提供
			// `status.Code` は gRPC のステータスコードを取り出す
			// see https://pkg.go.dev/google.golang.org/grpc/status
			if status.Code(err) == codes.Canceled {
				// キャンセル終了ならループを脱出
				break
			}
			return fmt.Errorf("receiving boot response: %w", err)
		}
		fmt.Printf("Boot: %s\n", resp.Message)
	}

	return nil
}
