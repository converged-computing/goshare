package main

import (
	"context"
	"flag"
	"io"
	"log"
	"math/rand"
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

	rand.Seed(time.Now().Unix())
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

	// Send and receive, providing the command, until we close
	client := pb.NewStreamClient(conn)
	stream, err := client.Command(context.Background())

	ctx := stream.Context()
	done := make(chan bool)
	var pid int32

	// first goroutine sends command to stream, and expects a pid back
	go func() {

		// The message includes the command (could eventually include other things)
		message := pb.CommandRequest{Command: cmd}
		if err := stream.Send(&message); err != nil {
			log.Fatalf("can not send %v", err)
		}
		log.Printf("%d sent command", message.Command)

		// Short sleep to get pid back
		time.Sleep(time.Millisecond * 200)

		// Expect to receive a pid back
		resp, err := stream.Recv()
		if err == io.EOF {
			close(done)
			return
		}
		if err != nil {
			log.Fatalf("can not receive %v", err)
		}
		pid = resp.Pid
		log.Printf("new pid %d received", pid)

		// Close our stream here
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine expects a finished response back.
	// if stream is finished it closes done channel
	/*go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			pid := resp.Result
			log.Printf("new max %d received", pid)
		}
	}()*/

	// last goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Printf("finished with pid=%d", pid)
}
