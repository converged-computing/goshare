// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: proto/echo.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	Echo_Echo_FullMethodName = "/echo.Echo/Echo"
)

// EchoClient is the client API for Echo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EchoClient interface {
	Echo(ctx context.Context, in *EchoMessage, opts ...grpc.CallOption) (*EchoResponse, error)
}

type echoClient struct {
	cc grpc.ClientConnInterface
}

func NewEchoClient(cc grpc.ClientConnInterface) EchoClient {
	return &echoClient{cc}
}

func (c *echoClient) Echo(ctx context.Context, in *EchoMessage, opts ...grpc.CallOption) (*EchoResponse, error) {
	out := new(EchoResponse)
	err := c.cc.Invoke(ctx, Echo_Echo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// EchoServer is the server API for Echo service.
// All implementations should embed UnimplementedEchoServer
// for forward compatibility
type EchoServer interface {
	Echo(context.Context, *EchoMessage) (*EchoResponse, error)
}

// UnimplementedEchoServer should be embedded to have forward compatible implementations.
type UnimplementedEchoServer struct {
}

func (UnimplementedEchoServer) Echo(context.Context, *EchoMessage) (*EchoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Echo not implemented")
}

// UnsafeEchoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EchoServer will
// result in compilation errors.
type UnsafeEchoServer interface {
	mustEmbedUnimplementedEchoServer()
}

func RegisterEchoServer(s grpc.ServiceRegistrar, srv EchoServer) {
	s.RegisterService(&Echo_ServiceDesc, srv)
}

func _Echo_Echo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EchoMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EchoServer).Echo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Echo_Echo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EchoServer).Echo(ctx, req.(*EchoMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// Echo_ServiceDesc is the grpc.ServiceDesc for Echo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Echo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "echo.Echo",
	HandlerType: (*EchoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Echo",
			Handler:    _Echo_Echo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/echo.proto",
}
