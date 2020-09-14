package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/datastore/proto"
)

//Read reads out some data
func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	return nil, fmt.Errorf("Not implemented yet")
}

//Write writes out a key
func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
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
