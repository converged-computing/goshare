package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/converged-computing/goshare/internal/pb"
	"github.com/converged-computing/goshare/pkg/service"
	"google.golang.org/grpc"
)

const (
	protocol = "unix"
)

var (
	l = log.New(os.Stderr, "🟦️ service: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {
	sock := flag.String("s", "", "path to socket")
	help := flag.Bool("h", false, "usage help")
	flag.Parse()

	// This won't work if the filesystem isn't shared heres
	sockAddr := *sock
	if sockAddr == "" {
		sockAddr = "/tmp/goshare.sock"
	}
	if *help {
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "[-s path.socket] /path.socket")
		flag.PrintDefaults()
		os.Exit(0)
	}
	l.Printf("starting service at socket %s\n", sockAddr)
	if _, err := os.Stat(sockAddr); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockAddr); err != nil {
			l.Fatal(err)
		}
	}

	listener, err := net.Listen(protocol, sockAddr)
	if err != nil {
		log.Fatal(err)
	}

	l.Printf("creating a new service to listen at %s\n", sockAddr)
	s := grpc.NewServer()
	pb.RegisterStreamServer(s, &service.Server{})

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
