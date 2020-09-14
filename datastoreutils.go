package main

import (
	"fmt"

	pb "github.com/brotherlogic/datastore/proto"
	"golang.org/x/net/context"
)

func (s *Server) saveToWriteLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badWrite {
		return fmt.Errorf("Nope")
	}
	return nil
}

func (s *Server) saveToFanoutLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badFanout {
		return fmt.Errorf("Nope")
	}
	return nil
}
