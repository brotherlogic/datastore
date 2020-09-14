package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/datastore/proto"
)

//Read reads out some data
func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	dir, file := extractFilename(req.GetKey())
	resp, err := s.readFile(dir, file)

	code := status.Convert(err)
	if code.Code() != codes.OK && code.Code() != codes.NotFound {
		return nil, err
	}

	//Fast return path - we have the key *and* it's in the cache (i.e. our version is fresh)
	if code.Code() == codes.OK && s.cachedKey[req.GetKey()] {
		return &pb.ReadResponse{
			Value:     resp.GetValue(),
			Timestamp: resp.GetTimestamp(),
			Origin:    s.Registry.GetIdentifier(),
		}, nil
	}

	return nil, fmt.Errorf("Not implemented yet")
}

//Write writes out a key
func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	// Keys should not start with a '/'
	if strings.HasPrefix(req.GetKey(), "/") {
		return nil, fmt.Errorf("Keys should not start with a backslash: %v", req.GetKey())
	}

	internal := &pb.WriteInternalRequest{
		Key:       req.GetKey(),
		Value:     req.GetValue(),
		Timestamp: time.Now().Unix(),
		Origin:    s.Registry.GetIdentifier(),
	}

	err := s.saveToWriteLog(ctx, internal)
	if err != nil {
		return nil, err
	}

	err = s.saveToFanoutLog(ctx, internal)
	if err != nil {
		return nil, err
	}

	return &pb.WriteResponse{NewVersion: internal.GetTimestamp()}, nil
}

//WriteInternal is a fanout write
func (s *Server) WriteInternal(ctx context.Context, req *pb.WriteInternalRequest) (*pb.WriteInternalResponse, error) {
	err := s.saveToWriteLog(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.WriteInternalResponse{}, nil
}
