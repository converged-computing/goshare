// gRPC with Unix Domain Socket example (server side)
//
// # REFERENCES
// 	- https://qiita.com/marnie_ms4/items/4582a1a0db363fe246f3
// 	- http://yamahiro0518.hatenablog.com/entry/2016/02/01/215908
// 	- https://zenn.dev/hsaki/books/golang-grpc-starting/viewer/client
// 	- https://stackoverflow.com/a/46279623
// 	- https://stackoverflow.com/a/18479916
//	- https://qiita.com/hnakamur/items/848097aad846d40ae84b

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/converged-computing/goshare/internal/pb"
	"github.com/converged-computing/goshare/internal/service"
	"google.golang.org/grpc"
)

const (
	protocol = "unix"
)

func main() {
	sock := flag.String("s", "", "path to socket")
	help := flag.Bool("h", false, "usage help")
	flag.Parse()

	// This won't work if the filesystem isn't shared heres
	sockAddr := *sock
	if sockAddr == "" {
		sockAddr = "/tmp/echo.sock"
	}

	if *help {
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "[-s path.socket] /path.socket")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if _, err := os.Stat(sockAddr); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockAddr); err != nil {
			log.Fatal(err)
		}
	}

	listener, err := net.Listen(protocol, sockAddr)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	echo := service.NewEchoService()

	pb.RegisterEchoServer(server, echo)

	server.Serve(listener)
}
