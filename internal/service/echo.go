package service

import (
	"context"
	"strings"

	"github.com/converged-computing/goshare/internal/pb"
)

type EchoServiceImpl struct{}

var _ pb.EchoServer = (*EchoServiceImpl)(nil)

func NewEchoService() pb.EchoServer {
	return new(EchoServiceImpl)
}

func (e *EchoServiceImpl) Echo(ctx context.Context, message *pb.EchoMessage) (*pb.EchoResponse, error) {
	s := strings.ToUpper(message.Data)
	r := &pb.EchoResponse{
		Data: s,
	}

	return r, nil
}
