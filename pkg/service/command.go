package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/converged-computing/goshare/internal/pb"
	"github.com/converged-computing/goshare/lib/command"
)

type Server struct {
	pb.UnimplementedStreamServer
}

var (
	// Log with prefix for service
	l = log.New(os.Stderr, "🟦️ service: ", log.Ldate|log.Ltime|log.Lshortfile)
)

// Command is a service Endpoint for a streaming (more interactive) response
func (s Server) Command(srv pb.Stream_CommandServer) error {

	l.Println("start new stream request")
	ctx := srv.Context()

	for {

		// exit if context is done, otherwise continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream, typically the command from the client
		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			l.Println("exit, we received an end of file")
			return nil
		}

		// Some error that we don't expect
		if err != nil {
			l.Printf("received error %v", err)
			continue
		}

		// If we receive a Command, we need to execute it
		if req.Command != "" {
			l.Printf("Received command %s", req.Command)

			// Run the command and return a response back
			res, wrapper, err := runCommand(req)
			if err != nil {
				l.Printf("received error %v", err)
				continue
			}

			// Send response back to client
			if err := srv.Send(res); err != nil {
				l.Printf("sending back response error %v", err)
			}
			l.Printf("send new pid=%d", res.Pid)

			// Wait for the command to finish and return done!
			l.Printf("Process started with PID: %d\n", wrapper.Command.Process.Pid)
			err = wrapper.Command.Wait()

			// Update the res with the output
			output := wrapper.Builder.String()
			if output != "" {
				l.Printf("send final output: %s", output)
			}
			res.Output = output
			res.Done = 1

			if err != nil {
				l.Printf("Error waiting for command: %v\n", err)
			}

			// Send the response that is now completed
			if err := srv.Send(res); err != nil {
				l.Printf("sending final response error %v", err)
			}
			return ctx.Err()
		}
	}
}

// Service endpoint to receive a command, execute, and return the pid
func runCommand(message *pb.CommandRequest) (*pb.CommandResponse, command.CommandWrapper, error) {

	// This returns back the command so we can get the pid, wait on it, etc.
	wrapper, err := command.RunDetachedCommand(strings.Split(message.Command, " "), []string{})
	var r pb.CommandResponse
	if err != nil {
		errorPid := int32(-1)
		r = pb.CommandResponse{
			Pid:        errorPid,
			Error:      fmt.Sprintf("%x", err),
			Returncode: int32(-1),
		}
	} else {
		r = pb.CommandResponse{
			Pid:        int32(wrapper.Command.Process.Pid),
			Error:      "",
			Returncode: int32(0),
		}
	}
	return &r, wrapper, nil
}
