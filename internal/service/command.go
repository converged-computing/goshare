package service

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/converged-computing/goshare/internal/pb"
	"github.com/converged-computing/goshare/lib/command"
	_ "google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedStreamServer
}

// Command is a service Endpoint for a streaming (more interactive) response
func (s Server) Command(srv pb.Stream_CommandServer) error {

	log.Println("start new stream request")
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
			log.Println("exit, we received an end of file")
			return nil
		}

		// Some error that we don't expect
		if err != nil {
			log.Printf("received error %v", err)
			continue
		}

		// If we receive a Command, we need to execute it
		if req.Command != "" {
			log.Printf("Received command %s", req.Command)

			// Run the command and return a response back
			res, cmd, err := runCommand(req)
			if err != nil {

				// TODO custom logic here...
				log.Printf("received error %v", err)
				continue
			}

			// Send response back to client
			if err := srv.Send(res); err != nil {
				log.Printf("sending back response error %v", err)
			}
			log.Printf("send new pid=%d", res.Pid)

			// Wait for the command to finish and return done!
			log.Printf("Process started with PID: %d\n", cmd.Process.Pid)
			err = cmd.Wait()
			if err != nil {
				fmt.Printf("Error waiting for command: %v\n", err)
			}
			return ctx.Err()
		}
	}
}

// Service endpoint to receive a command, execute, and return the pid
func runCommand(message *pb.CommandRequest) (*pb.CommandResponse, *exec.Cmd, error) {

	// This returns back the command so we can get the pid, wait on it, etc.
	cmd, err := command.RunDetachedCommand(strings.Split(message.Command, " "), []string{})
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
			Pid:        int32(cmd.Process.Pid),
			Error:      "",
			Returncode: int32(0),
		}
	}
	return &r, cmd, nil
}
