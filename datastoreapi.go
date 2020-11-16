package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/datastore/proto"
)

func min(a int32, b int) int {
	if int(a) > b {
		return b
	}
	return int(a)
}

func max(a int32, b int) int32 {
	if int(a) < b {
		return int32(b)
	}
	return a
}

func (s *Server) buildConsensus(ctx context.Context, key string, consensus int32, save bool) (*pb.ReadResponse, error) {
	// Default time to build consensus is 1 minute - but knock one second off of the existing context if it has a timeout
	newTime := time.Minute
	deadline, ok := ctx.Deadline()
	if ok {
		newTime = deadline.Sub(time.Now().Add(-time.Second))
	}
	nctx, cancel := context.WithTimeout(ctx, newTime)
	defer cancel()

	waitGroup := &sync.WaitGroup{}
	reads := []*pb.ReadResponse{}
	friends := s.getFriends(ctx)
	for _, friend := range friends[0:min(consensus, len(friends))] {
		//Only wait for full consensus
		waitGroup.Add(1)
		go func() {
			read, err := s.remoteRead(nctx, friend, key)
			if err == nil {
				reads = append(reads, read)
			}
			waitGroup.Done()
		}()
	}

	//Wait for all to complete (or at least timeout)
	waitGroup.Wait()

	//Find the best response
	var best *pb.ReadResponse
	mTime := int64(0)
	for _, read := range reads {
		if read.GetTimestamp() > mTime {
			mTime = read.GetTimestamp()
			best = read
		}
	}

	if best == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Could not find %v, even with %v friends", key, len(s.getFriends(ctx)))
	}

	// Save the consensus if we need to
	if save {
		// We can save it to the write log without a fanout
		s.saveToWriteLog(ctx,
			&pb.WriteInternalRequest{
				Key:       key,
				Value:     best.GetValue(),
				Timestamp: best.GetTimestamp(),
				Origin:    best.GetOrigin(),
			})
	}

	return best, nil
}

//Read reads out some data
func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	dir, file := extractFilename(req.GetKey())
	resp, err := s.readFile(dir, file)

	code := status.Convert(err)
	if code.Code() != codes.OK && code.Code() != codes.NotFound {
		return nil, err
	}

	//Fast return path - we have the key *and* it's in the cache (i.e. our version is fresh), or we explicitly asked for the local version
	s.Log(fmt.Sprintf("CODE: %v, Cache: %v", code.Code(), s.cachedKey))
	if code.Code() == codes.OK && (s.cachedKey[req.GetKey()] || req.GetConsensus() == 0) {
		return &pb.ReadResponse{
			Value:     resp.GetValue(),
			Timestamp: resp.GetTimestamp(),
			Origin:    s.Registry.GetIdentifier(),
		}, nil
	}

	// If we're not asking for consensus, return not found
	if req.GetConsensus() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Could not find %v", req)
	}

	//Let's get a consensus on the latest
	return s.buildConsensus(ctx, req.GetKey(), req.GetConsensus(), !s.cachedKey[req.GetKey()])
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
