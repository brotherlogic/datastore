// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// DatastoreServiceClient is the client API for DatastoreService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DatastoreServiceClient interface {
	Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error)
	Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error)
}

type datastoreServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDatastoreServiceClient(cc grpc.ClientConnInterface) DatastoreServiceClient {
	return &datastoreServiceClient{cc}
}

func (c *datastoreServiceClient) Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error) {
	out := new(ReadResponse)
	err := c.cc.Invoke(ctx, "/datastore.DatastoreService/Read", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *datastoreServiceClient) Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error) {
	out := new(WriteResponse)
	err := c.cc.Invoke(ctx, "/datastore.DatastoreService/Write", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DatastoreServiceServer is the server API for DatastoreService service.
// All implementations should embed UnimplementedDatastoreServiceServer
// for forward compatibility
type DatastoreServiceServer interface {
	Read(context.Context, *ReadRequest) (*ReadResponse, error)
	Write(context.Context, *WriteRequest) (*WriteResponse, error)
}

// UnimplementedDatastoreServiceServer should be embedded to have forward compatible implementations.
type UnimplementedDatastoreServiceServer struct {
}

func (UnimplementedDatastoreServiceServer) Read(context.Context, *ReadRequest) (*ReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedDatastoreServiceServer) Write(context.Context, *WriteRequest) (*WriteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Write not implemented")
}

// UnsafeDatastoreServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DatastoreServiceServer will
// result in compilation errors.
type UnsafeDatastoreServiceServer interface {
	mustEmbedUnimplementedDatastoreServiceServer()
}

func RegisterDatastoreServiceServer(s grpc.ServiceRegistrar, srv DatastoreServiceServer) {
	s.RegisterService(&_DatastoreService_serviceDesc, srv)
}

func _DatastoreService_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatastoreServiceServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/datastore.DatastoreService/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatastoreServiceServer).Read(ctx, req.(*ReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DatastoreService_Write_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatastoreServiceServer).Write(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/datastore.DatastoreService/Write",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatastoreServiceServer).Write(ctx, req.(*WriteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DatastoreService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "datastore.DatastoreService",
	HandlerType: (*DatastoreServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Read",
			Handler:    _DatastoreService_Read_Handler,
		},
		{
			MethodName: "Write",
			Handler:    _DatastoreService_Write_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "datastore.proto",
}

// DatastoreInternalServiceClient is the client API for DatastoreInternalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DatastoreInternalServiceClient interface {
	WriteInternal(ctx context.Context, in *WriteInternalRequest, opts ...grpc.CallOption) (*WriteInternalResponse, error)
}

type datastoreInternalServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDatastoreInternalServiceClient(cc grpc.ClientConnInterface) DatastoreInternalServiceClient {
	return &datastoreInternalServiceClient{cc}
}

func (c *datastoreInternalServiceClient) WriteInternal(ctx context.Context, in *WriteInternalRequest, opts ...grpc.CallOption) (*WriteInternalResponse, error) {
	out := new(WriteInternalResponse)
	err := c.cc.Invoke(ctx, "/datastore.DatastoreInternalService/WriteInternal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DatastoreInternalServiceServer is the server API for DatastoreInternalService service.
// All implementations should embed UnimplementedDatastoreInternalServiceServer
// for forward compatibility
type DatastoreInternalServiceServer interface {
	WriteInternal(context.Context, *WriteInternalRequest) (*WriteInternalResponse, error)
}

// UnimplementedDatastoreInternalServiceServer should be embedded to have forward compatible implementations.
type UnimplementedDatastoreInternalServiceServer struct {
}

func (UnimplementedDatastoreInternalServiceServer) WriteInternal(context.Context, *WriteInternalRequest) (*WriteInternalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WriteInternal not implemented")
}

// UnsafeDatastoreInternalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DatastoreInternalServiceServer will
// result in compilation errors.
type UnsafeDatastoreInternalServiceServer interface {
	mustEmbedUnimplementedDatastoreInternalServiceServer()
}

func RegisterDatastoreInternalServiceServer(s grpc.ServiceRegistrar, srv DatastoreInternalServiceServer) {
	s.RegisterService(&_DatastoreInternalService_serviceDesc, srv)
}

func _DatastoreInternalService_WriteInternal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteInternalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DatastoreInternalServiceServer).WriteInternal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/datastore.DatastoreInternalService/WriteInternal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DatastoreInternalServiceServer).WriteInternal(ctx, req.(*WriteInternalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DatastoreInternalService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "datastore.DatastoreInternalService",
	HandlerType: (*DatastoreInternalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WriteInternal",
			Handler:    _DatastoreInternalService_WriteInternal_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "datastore.proto",
}