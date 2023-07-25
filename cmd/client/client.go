// gRPC with Unix Domain Socket example (client side)
//
// # REFERENCES
// 	- https://qiita.com/marnie_ms4/items/4582a1a0db363fe246f3
// 	- http://yamahiro0518.hatenablog.com/entry/2016/02/01/215908
// 	- https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/client
// 	- https://stackoverflow.com/a/46279623
// 	- https://stackoverflow.com/a/18479916

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/converged-computing/goshare/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	protocol = "unix"
)

func main() {

	sock := flag.String("s", "", "path to socket")
	command := flag.String("c", "", "command")
	flag.Parse()

	// This won't work if the filesystem isn't shared heres
	sockAddr := *sock
	cmd := *command
	if sockAddr == "" {
		sockAddr = "/tmp/echo.sock"
	}

	// Testing command
	if cmd == "" {
		cmd = "echo hello world"
	}

	var (
		rootCtx          = context.Background()
		mainCtx, mainCxl = context.WithCancel(rootCtx)
	)
	defer mainCxl()

	//
	// Connect
	//
	var (
		credentials = insecure.NewCredentials() // No SSL/TLS
		dialer      = func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, protocol, addr)
		}
		options = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials),
			grpc.WithBlock(),
			grpc.WithContextDialer(dialer),
		}
	)

	conn, err := grpc.Dial(sockAddr, options...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//
	// Send & Recv
	//
	var (
		client = pb.NewCommandClient(conn)
	)

	func() {
		ctx, cancel := context.WithTimeout(mainCtx, 1*time.Second)
		defer cancel()

		message := pb.CommandRequest{Command: cmd}
		res, err := client.Command(ctx, &message)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(res)
	}()
}
