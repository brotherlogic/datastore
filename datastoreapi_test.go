package main

import (
	"context"
	"os"
	"testing"

	pb "github.com/brotherlogic/datastore/proto"
	dpb "github.com/brotherlogic/discovery/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func InitTest(remove bool, dir string) *Server {
	s := Init()
	s.SkipLog = true
	s.basepath = dir
	if remove {
		os.RemoveAll(dir)
	}
	s.Registry = &dpb.RegistryEntry{Identifier: "test"}
	return s
}

func TestReadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}
	_, err = s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err == nil {
		t.Fatalf("Bad read: %v", err)
	}

}

func TestMultiWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	var err error
	for i := 0; i < 200; i++ {
		_, err = s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	}

	// We shouldn't be able to do 200 writes when the fanout isn't running
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestBadKey(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "/testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestBadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badWrite = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestBadMarshal(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badMarshal = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestBadFanout(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badFanout = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}
