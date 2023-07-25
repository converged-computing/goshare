package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/converged-computing/goshare/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	protocol = "unix"
)

var sockAddr string

func main() {

	// TODO how to do this now?
	//	rand.NewRand(rand.NewSource(rand.Seed(time.Now().Unix())))
	flag.StringVar(&sockAddr, "s", "/tmp/echo.sock", "path to socket")
	flag.Parse()
	listcmd := flag.Args()

	// Testing command
	if len(listcmd) == 0 {
		listcmd = []string{"echo", "hello", "world"}
	}

	// Serialized as a string
	cmd := strings.Join(listcmd, " ")
	log.Printf("      socket path: %s", sockAddr)
	log.Printf("requested command: %s", cmd)

	// Connection options
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
	if err != nil {
		log.Fatalf("can not issue command to stream client %v", err)
	}

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
		log.Printf("%s sent command", message.Command)

		// Short sleep to get pid back
		time.Sleep(time.Millisecond * 200)

		// Expect to receive a pid back
		resp, err := stream.Recv()
		if err == io.EOF {
			log.Printf("end of file received, closing")
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
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				log.Printf("end of file received, closing")
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			// TODO get back output here?
			log.Printf("new output received? %x", resp)
		}
	}()

	// last goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
	}()

	<-done
	log.Printf("finished with client request")
}
