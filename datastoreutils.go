package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/datastore/proto"
)

func (s *Server) writeToDir(ctx context.Context, dir, file string, req *pb.WriteInternalRequest) error {
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	os.MkdirAll(dir, 0777)
	return ioutil.WriteFile(dir+file, data, 0644)
}

func (s *Server) saveToWriteLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badWrite {
		return fmt.Errorf("Nope")
	}

	// Fail the write if the queue if fulll
	if len(s.writeQueue) > 100 {
		return status.Errorf(codes.ResourceExhausted, "The write queue is full")
	}

	writeLogDir := s.basepath + "internal/towrite/"
	filename := fmt.Sprintf("%v.%v", strings.Replace(req.GetKey(), "/", "-", -1), req.GetTimestamp())
	err := s.writeToDir(ctx, writeLogDir, filename, req)
	if err != nil {
		return err
	}

	s.writeQueue <- filename

	return nil
}

func (s *Server) saveToFanoutLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badFanout {
		return fmt.Errorf("Nope")
	}
	return nil
}
