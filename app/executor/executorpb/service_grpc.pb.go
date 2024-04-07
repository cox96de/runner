// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: service.proto

package executorpb

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
	Executor_GetRuntimeInfo_FullMethodName = "/Executor/GetRuntimeInfo"
	Executor_Ping_FullMethodName           = "/Executor/Ping"
	Executor_Environment_FullMethodName    = "/Executor/Environment"
	Executor_StartCommand_FullMethodName   = "/Executor/StartCommand"
	Executor_WaitCommand_FullMethodName    = "/Executor/WaitCommand"
	Executor_GetCommandLog_FullMethodName  = "/Executor/GetCommandLog"
)

// ExecutorClient is the client API for Executor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExecutorClient interface {
	GetRuntimeInfo(ctx context.Context, in *GetRuntimeInfoRequest, opts ...grpc.CallOption) (*GetRuntimeInfoResponse, error)
	// Ping is used to check if the executor is alive.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// Environment returns the environment variables of the executor.
	// Just like the os.Environ() function in Go.
	Environment(ctx context.Context, in *EnvironmentRequest, opts ...grpc.CallOption) (*EnvironmentResponse, error)
	// StartCommand starts a command in the executor.
	StartCommand(ctx context.Context, in *StartCommandRequest, opts ...grpc.CallOption) (*StartCommandResponse, error)
	// WaitCommand waits for a command to finish.
	WaitCommand(ctx context.Context, in *WaitCommandRequest, opts ...grpc.CallOption) (*WaitCommandResponse, error)
	// GetCommandLog returns the log of a command.
	GetCommandLog(ctx context.Context, in *GetCommandLogRequest, opts ...grpc.CallOption) (Executor_GetCommandLogClient, error)
}

type executorClient struct {
	cc grpc.ClientConnInterface
}

func NewExecutorClient(cc grpc.ClientConnInterface) ExecutorClient {
	return &executorClient{cc}
}

func (c *executorClient) GetRuntimeInfo(ctx context.Context, in *GetRuntimeInfoRequest, opts ...grpc.CallOption) (*GetRuntimeInfoResponse, error) {
	out := new(GetRuntimeInfoResponse)
	err := c.cc.Invoke(ctx, Executor_GetRuntimeInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, Executor_Ping_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorClient) Environment(ctx context.Context, in *EnvironmentRequest, opts ...grpc.CallOption) (*EnvironmentResponse, error) {
	out := new(EnvironmentResponse)
	err := c.cc.Invoke(ctx, Executor_Environment_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorClient) StartCommand(ctx context.Context, in *StartCommandRequest, opts ...grpc.CallOption) (*StartCommandResponse, error) {
	out := new(StartCommandResponse)
	err := c.cc.Invoke(ctx, Executor_StartCommand_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorClient) WaitCommand(ctx context.Context, in *WaitCommandRequest, opts ...grpc.CallOption) (*WaitCommandResponse, error) {
	out := new(WaitCommandResponse)
	err := c.cc.Invoke(ctx, Executor_WaitCommand_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorClient) GetCommandLog(ctx context.Context, in *GetCommandLogRequest, opts ...grpc.CallOption) (Executor_GetCommandLogClient, error) {
	stream, err := c.cc.NewStream(ctx, &Executor_ServiceDesc.Streams[0], Executor_GetCommandLog_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &executorGetCommandLogClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Executor_GetCommandLogClient interface {
	Recv() (*Log, error)
	grpc.ClientStream
}

type executorGetCommandLogClient struct {
	grpc.ClientStream
}

func (x *executorGetCommandLogClient) Recv() (*Log, error) {
	m := new(Log)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ExecutorServer is the server API for Executor service.
// All implementations must embed UnimplementedExecutorServer
// for forward compatibility
type ExecutorServer interface {
	GetRuntimeInfo(context.Context, *GetRuntimeInfoRequest) (*GetRuntimeInfoResponse, error)
	// Ping is used to check if the executor is alive.
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// Environment returns the environment variables of the executor.
	// Just like the os.Environ() function in Go.
	Environment(context.Context, *EnvironmentRequest) (*EnvironmentResponse, error)
	// StartCommand starts a command in the executor.
	StartCommand(context.Context, *StartCommandRequest) (*StartCommandResponse, error)
	// WaitCommand waits for a command to finish.
	WaitCommand(context.Context, *WaitCommandRequest) (*WaitCommandResponse, error)
	// GetCommandLog returns the log of a command.
	GetCommandLog(*GetCommandLogRequest, Executor_GetCommandLogServer) error
	mustEmbedUnimplementedExecutorServer()
}

// UnimplementedExecutorServer must be embedded to have forward compatible implementations.
type UnimplementedExecutorServer struct {
}

func (UnimplementedExecutorServer) GetRuntimeInfo(context.Context, *GetRuntimeInfoRequest) (*GetRuntimeInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRuntimeInfo not implemented")
}
func (UnimplementedExecutorServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedExecutorServer) Environment(context.Context, *EnvironmentRequest) (*EnvironmentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Environment not implemented")
}
func (UnimplementedExecutorServer) StartCommand(context.Context, *StartCommandRequest) (*StartCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartCommand not implemented")
}
func (UnimplementedExecutorServer) WaitCommand(context.Context, *WaitCommandRequest) (*WaitCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitCommand not implemented")
}
func (UnimplementedExecutorServer) GetCommandLog(*GetCommandLogRequest, Executor_GetCommandLogServer) error {
	return status.Errorf(codes.Unimplemented, "method GetCommandLog not implemented")
}
func (UnimplementedExecutorServer) mustEmbedUnimplementedExecutorServer() {}

// UnsafeExecutorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExecutorServer will
// result in compilation errors.
type UnsafeExecutorServer interface {
	mustEmbedUnimplementedExecutorServer()
}

func RegisterExecutorServer(s grpc.ServiceRegistrar, srv ExecutorServer) {
	s.RegisterService(&Executor_ServiceDesc, srv)
}

func _Executor_GetRuntimeInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRuntimeInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServer).GetRuntimeInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Executor_GetRuntimeInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServer).GetRuntimeInfo(ctx, req.(*GetRuntimeInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Executor_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Executor_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Executor_Environment_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EnvironmentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServer).Environment(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Executor_Environment_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServer).Environment(ctx, req.(*EnvironmentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Executor_StartCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServer).StartCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Executor_StartCommand_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServer).StartCommand(ctx, req.(*StartCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Executor_WaitCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WaitCommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServer).WaitCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Executor_WaitCommand_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServer).WaitCommand(ctx, req.(*WaitCommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Executor_GetCommandLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetCommandLogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ExecutorServer).GetCommandLog(m, &executorGetCommandLogServer{stream})
}

type Executor_GetCommandLogServer interface {
	Send(*Log) error
	grpc.ServerStream
}

type executorGetCommandLogServer struct {
	grpc.ServerStream
}

func (x *executorGetCommandLogServer) Send(m *Log) error {
	return x.ServerStream.SendMsg(m)
}

// Executor_ServiceDesc is the grpc.ServiceDesc for Executor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Executor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Executor",
	HandlerType: (*ExecutorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRuntimeInfo",
			Handler:    _Executor_GetRuntimeInfo_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _Executor_Ping_Handler,
		},
		{
			MethodName: "Environment",
			Handler:    _Executor_Environment_Handler,
		},
		{
			MethodName: "StartCommand",
			Handler:    _Executor_StartCommand_Handler,
		},
		{
			MethodName: "WaitCommand",
			Handler:    _Executor_WaitCommand_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetCommandLog",
			Handler:       _Executor_GetCommandLog_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "service.proto",
}
