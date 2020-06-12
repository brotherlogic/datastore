package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/datastore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func (s *Server) getLatest(key string) (string, int64, error) {
	if strings.HasPrefix(key, "/") {
		key = key[1:]
	}
	if !strings.HasSuffix(key, "/") {
		key = key + "/"
	}
	fullpath := s.basepath + key
	bestv := 0

	files, err := ioutil.ReadDir(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			err = status.Errorf(codes.NotFound, "%v", err)
		}
		return fullpath, 0, err
	}
	for _, f := range files {
		v, _ := strconv.Atoi(f.Name())
		if v > bestv {
			bestv = v
		}
	}

	return fullpath, int64(bestv), nil
}

//Read reads out some data
func (s *Server) Read(ctx context.Context, req *pb.ReadRequest) (*pb.ReadResponse, error) {
	path, version, err := s.getLatest(req.GetKey())
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%v%v", path, version))
	if err != nil || s.testingfails {
		return nil, status.Errorf(codes.Internal, "Error reading file (%v) - >%v", path, err)
	}

	info, _ := os.Stat(fmt.Sprintf("%v%v", path, version))
	return &pb.ReadResponse{Value: &google_protobuf.Any{Value: data}, Version: version, Timestamp: info.ModTime().Unix()}, nil
}

//Write writes out a key
func (s *Server) Write(ctx context.Context, req *pb.WriteRequest) (*pb.WriteResponse, error) {
	path, version, err := s.getLatest(req.GetKey())
	c := status.Convert(err).Code()
	if (c != codes.NotFound && c != codes.OK) || s.testingfails {
		if s.testingfails {
			err = fmt.Errorf("Built to fail")
		}
		return nil, err
	}

	nversion := version + 1
	npath := fmt.Sprintf("%v%v", path, nversion)

	os.MkdirAll(path, 0777)
	err = ioutil.WriteFile(npath, req.GetValue().GetValue(), 0644)

	// Fanout the write if needed
	if req.GetFanoutMinimum() >= 0 {
		key := fmt.Sprintf("%v", time.Now().UnixNano())
		s.fanout(req, key, int(req.GetFanoutMinimum()))
	}

	return &pb.WriteResponse{NewVersion: nversion}, err
}

//Friend befriends another server
func (s *Server) Friend(ctx context.Context, req *pb.FriendRequest) (*pb.FriendResponse, error) {
	defer Friends.Set(float64(len(s.friends)))
	for _, infriend := range req.GetFriend() {
		found := false
		for _, friend := range s.friends {
			if friend == infriend {
				found = true
			}
		}

		if !found {
			s.friends = append(s.friends, infriend)
		}
	}

	return &pb.FriendResponse{Friend: append(s.friends, fmt.Sprintf("%v:%v", s.Registry.Identifier, s.Registry.Port))}, nil
}
