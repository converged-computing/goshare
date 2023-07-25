package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/converged-computing/goshare/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	protocol = "unix"
)

var (
	l               = log.New(os.Stderr, "🟪️  client: ", log.Ldate|log.Ltime|log.Lshortfile)
	sockAddr string = "/tmp/echo.sock"
)

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
	l.Printf("      socket path: %s", sockAddr)
	l.Printf("requested command: %s", cmd)

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
		l.Fatalf("can not issue command to stream client %v", err)
	}

	ctx := stream.Context()
	done := make(chan bool)
	var pid int32

	// first goroutine sends command to stream, and expects a pid back
	go func() {

		// The message includes the command (could eventually include other things)
		message := pb.CommandRequest{Command: cmd}
		if err := stream.Send(&message); err != nil {
			l.Fatalf("can not send %v", err)
		}
		l.Printf("%s sent command", message.Command)

		// Short sleep to get pid back
		time.Sleep(time.Millisecond * 200)

		// Expect to receive a pid back
		resp, err := stream.Recv()
		if err == io.EOF {
			l.Printf("end of file received, closing")
			close(done)
			return
		}
		if err != nil {
			l.Fatalf("can not receive %v", err)
		}
		pid = resp.Pid
		l.Printf("new pid %d received", pid)

		// Close our stream here
		if err := stream.CloseSend(); err != nil {
			l.Println(err)
		}
	}()

	// second goroutine expects a finished response back.
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				l.Printf("end of file received, closing")
				close(done)
				return
			}
			if err != nil {
				l.Fatalf("can not receive %v", err)
			}
			if resp.Output != "" {
				l.Printf("new output received: %s", resp.Output)
			}
		}
	}()

	// last goroutine closes done channel
	go func() {
		<-ctx.Done()
	}()

	<-done
	l.Printf("finished with client request")
}
