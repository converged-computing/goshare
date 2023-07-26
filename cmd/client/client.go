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
	sockAddr string = "/tmp/goshare.sock"
	workdir  string = "/tmp"
)

func main() {

	flag.StringVar(&sockAddr, "s", "/tmp/goshare.sock", "path to socket")
	flag.StringVar(&workdir, "w", "", "working directory to run job")
	flag.Parse()
	listcmd := flag.Args()

	// Testing command
	if len(listcmd) == 0 {
		listcmd = []string{"echo", "hello", "world"}
	}

	// Serialized as a string
	cmd := strings.Join(listcmd, " ")
	l.Printf("socket path: %s\n", sockAddr)
	l.Printf("requested command: %s\n", cmd)

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

	// first goroutine sends command to send command
	go func() {

		// The message includes the command (could eventually include other things)
		message := pb.CommandRequest{Command: cmd, Workdir: workdir}
		if err := stream.Send(&message); err != nil {
			l.Fatalf("can not send %v", err)
		}
		l.Printf("sent command: %s\n", message.Command)

		// Short sleep to get pid back
		time.Sleep(time.Millisecond * 200)

		// Close our stream here
		l.Printf("closing send\n")
		if err := stream.CloseSend(); err != nil {
			l.Println(err)
		}
	}()

	// second goroutine expects the pid and then output
	go func() {
		for {

			// Expect to receive a pid back
			resp, err := stream.Recv()

			// We always get back a PID
			pid = resp.Pid
			l.Printf("pid %d is active\n", pid)

			// If we are done, we close
			if resp.Done == 1 {
				if resp.Output != "" {
					l.Printf("new output received: %s", resp.Output)
				}
				l.Printf("process is done, closing\n")
				close(done)
				return
			}
			if err != nil {
				l.Fatalf("can not receive %v\n", err)
			}
			if err == io.EOF {
				l.Printf("end of file received, closing")
				close(done)
				return
			}
			time.Sleep(time.Millisecond * 200)
		}
	}()

	// last goroutine closes done channel
	go func() {
		<-ctx.Done()
	}()

	<-done
	l.Printf("finished with client request")
}
