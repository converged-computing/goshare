package service

import (
	"context"
	"strings"

	"github.com/converged-computing/goshare/internal/pb"
	"github.com/converged-computing/goshare/lib/command"
)

type CommandServiceImpl struct{}

var _ pb.CommandServer = (*CommandServiceImpl)(nil)

func NewCommandService() pb.CommandServer {
	return new(CommandServiceImpl)
}

// Service endpoint to receive a command, execute, and return the pid
func (e *CommandServiceImpl) Command(ctx context.Context, message *pb.CommandRequest) (*pb.CommandResponse, error) {
	pid, err := command.RunDetachedCommand(strings.Split(message.Command, " "), []string{})
	var r pb.CommandResponse
	if err != nil {
		errorPid := int32(-1)
		r = pb.CommandResponse{
			Pid: errorPid,
		}
	} else {
		r = pb.CommandResponse{
			Pid: int32(pid),
		}
	}

	return &r, nil
}
