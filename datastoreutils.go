package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/datastore/proto"
	"github.com/brotherlogic/goserver/utils"
)

func extractFilename(key string) (string, string) {
	val := strings.LastIndex(key, "/")
	if val < 0 {
		return "", key
	}
	return key[0 : val+1], key[val+1:]
}

func (s *Server) deleteFile(dir, file string) {
	s.Log(fmt.Sprintf("Removing %v%v -> %v", dir, file, os.Remove(s.basepath+dir+file)))
}

func (s *Server) readFile(dir, file string) (*pb.WriteInternalRequest, error) {
	data, err := ioutil.ReadFile(s.basepath + dir + file)

	if err != nil || s.badRead || (s.badFanoutRead && dir == "internal/fanout/") {
		if s.badRead || (s.badFanoutRead && dir == "internal/fanout/") {
			err = fmt.Errorf("Built to fail")
		}

		if os.IsNotExist(err) {
			return nil, status.Errorf(codes.NotFound, err.Error())
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

func (s *Server) processFanoutQueue() {
	for file := range s.fanoutQueue {
		//Load the file
		req, err := s.readFile("internal/fanout/", file)

		// Dump the file if we can't read it (it'll get picked up again on the next re-write)
		if err != nil {
			s.badQueueProcess++
			s.RaiseIssue(fmt.Sprintf("Bad file read for %v", file), fmt.Sprintf("Read error: %v", err))
			continue
		}

		ctx, cancel := utils.ManualContext("dsfo", "dsfo", time.Second*10, true)
		err = s.fanout(ctx, req)
		cancel()

		if err != nil {
			s.Log(fmt.Sprintf("Unable to fanout the write: %v", err))
			s.fanoutQueue <- file
		} else {
			s.deleteFile("internal/fanout/", file)
		}

		//Don't overload here
		time.Sleep(time.Second * 10)
	}
}
func (s *Server) processWriteQueue() {
	for file := range s.writeQueue {
		//Load the file
		req, err := s.readFile("internal/towrite/", file)

		// Dump the file if we can't read it (it'll get picked up again on the next re-write)
		if err != nil {
			s.badQueueProcess++
			s.RaiseIssue(fmt.Sprintf("Bad file read for %v", file), fmt.Sprintf("Read error: %v", err))
			continue
		}

		dir, file := extractFilename(req.GetKey())
		filename := fmt.Sprintf("%v.%v", strings.Replace(req.GetKey(), "/", "-", -1), req.GetTimestamp())

		// This also suggests some corruption somewhere, so pick it up again on the re-write loop
		err = s.writeToDir(dir, filename, req)
		if err != nil {
			s.badQueueProcess++
			s.RaiseIssue(fmt.Sprintf("Bad write for %v", file), fmt.Sprintf("Write error: %v", err))
			continue
		}

		// If we've got here then everything's fine
		s.deleteFile("internal/towrite/", filename)
		s.cachedKey[req.GetKey()] = true

		// No overload
		time.Sleep(time.Second * 10)
	}
}

func (s *Server) writeToDir(dir, file string, req *pb.WriteInternalRequest) error {
	data, err := proto.Marshal(req)
	if err != nil ||
		s.badMarshal ||
		(s.badBaseMarshal && !strings.HasPrefix(dir, "internal")) ||
		(s.badFanoutWrite && strings.HasPrefix(dir, "internal/fanout")) {
		if s.badMarshal || (s.badBaseMarshal && !strings.HasPrefix(dir, "internal")) || s.badFanoutWrite {
			err = fmt.Errorf("Bad marshal (%v) %v, %v, %v", dir, s.badMarshal, s.badBaseMarshal, s.badFanoutWrite)
		}
		return err
	}

	os.MkdirAll(s.basepath+dir, 0777)
	return ioutil.WriteFile(fmt.Sprintf("%v%v%v", s.basepath, dir, file), data, 0644)
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

func (s *Server) getFriends(ctx context.Context) []string {
	if len(s.friends) == 0 {
		s.populateFriends(ctx)
	}
	return s.friends
}

func (s *Server) saveToFanoutLog(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.badFanout {
		return fmt.Errorf("Nope")
	}

	// Fail the write if the queue if fulll
	if len(s.writeQueue) > 99 {
		return status.Errorf(codes.ResourceExhausted, "The write queue is full")
	}

	writeLogDir := "internal/fanout/"

	for _, friend := range s.getFriends(ctx) {
		copy := proto.Clone(req).(*pb.WriteInternalRequest)
		copy.Destination = friend
		filename := fmt.Sprintf("%v-%v.%v", strings.Replace(copy.GetKey(), "/", "-", -1), copy.GetDestination(), copy.GetTimestamp())
		err := s.writeToDir(writeLogDir, filename, copy)
		if err != nil {
			return err
		}

		s.fanoutQueue <- filename
	}

	return nil
}
