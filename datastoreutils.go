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

func extractFilename(key string) (string, string) {
	val := strings.LastIndex(key, "/")
	if val < 0 {
		return "", key
	}
	return key[0:val], key[val+1:]
}

func (s *Server) deleteFile(dir, file string) {
	s.Log(fmt.Sprintf("Removing %v%v -> %v", dir, file, os.Remove(dir+file)))
}

func (s *Server) readFile(dir, file string) (*pb.WriteInternalRequest, error) {
	data, err := ioutil.ReadFile(s.basepath + dir + file)
	if err != nil || s.badRead {
		if s.badRead {
			err = fmt.Errorf("Built to fail")
		}
		return nil, err
	}

	result := &pb.WriteInternalRequest{}
	err = proto.Unmarshal(data, result)
	if err != nil || s.badUnmarshal {
		if s.badUnmarshal {
			err = fmt.Errorf("Built to fail")
		}
		return nil, err
	}

	return result, nil
}

func (s *Server) processWriteQueue() {
	for file := range s.writeQueue {
		//Load the file
		req, err := s.readFile("internal/towrite/", file)

		// Dump the file if we can't read it (it'll get picked up again on the next re-write
		if err != nil {
			s.badQueueProcess++
			s.RaiseIssue(fmt.Sprintf("Bad file read for %v", file), fmt.Sprintf("Read error: %v", err))
			continue
		}

		dir, file := extractFilename(req.GetKey())
		err = s.writeToDir(dir, file, req)
		if err != nil {
			s.badQueueProcess++
			s.RaiseIssue(fmt.Sprintf("Bad write for %v", file), fmt.Sprintf("Write error: %v", err))
			continue
		}

		// If we've got here then everything's fine
		s.deleteFile("internal/towrite/", file)
	}
}

func (s *Server) writeToDir(dir, file string, req *pb.WriteInternalRequest) error {
	data, err := proto.Marshal(req)
	if err != nil || s.badMarshal || (s.badBaseMarshal && !strings.HasPrefix(dir, "internal")) {
		if s.badMarshal || (s.badBaseMarshal && !strings.HasPrefix(dir, "internal")) {
			err = fmt.Errorf("Bad marshal")
		}
		return err
	}

	os.MkdirAll(s.basepath+dir, 0777)
	return ioutil.WriteFile(s.basepath+dir+file, data, 0644)
}

func (s *Server) saveToWriteLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badWrite {
		return fmt.Errorf("Nope")
	}

	// Fail the write if the queue if fulll
	if len(s.writeQueue) > 99 {
		return status.Errorf(codes.ResourceExhausted, "The write queue is full")
	}

	writeLogDir := "internal/towrite/"
	filename := fmt.Sprintf("%v.%v", strings.Replace(req.GetKey(), "/", "-", -1), req.GetTimestamp())
	err := s.writeToDir(writeLogDir, filename, req)
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
