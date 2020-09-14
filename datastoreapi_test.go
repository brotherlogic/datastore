package main

import (
	"context"
	"os"
	"testing"
	"time"

	pb "github.com/brotherlogic/datastore/proto"
	dpb "github.com/brotherlogic/discovery/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func InitTest(remove bool, dir string) *Server {
	s := Init()
	s.SkipLog = true
	s.SkipIssue = true
	s.basepath = dir
	if remove {
		os.RemoveAll(dir)
	}
	s.Registry = &dpb.RegistryEntry{Identifier: "test"}
	return s
}

func drainQueue(s *Server) {
	//Run the queue to process the write
	go s.processWriteQueue()
	for len(s.writeQueue) > 0 {
		time.Sleep(time.Second)
	}
	close(s.writeQueue)
}

func TestReadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)

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

func TestBadRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badRead = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)

	// Should trigger a bad queue process
	if s.badQueueProcess != 1 {
		t.Errorf("Queue was processed fine")
	}
}

func TestBadUnmarshal(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badUnmarshal = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)

	// Should trigger a bad queue process
	if s.badQueueProcess != 1 {
		t.Errorf("Queue was processed fine")
	}
}

func TestBadBaseWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badBaseMarshal = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)

	// Should trigger a bad queue process
	if s.badQueueProcess != 1 {
		t.Errorf("Queue was processed fine")
	}
}
