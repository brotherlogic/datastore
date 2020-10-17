package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
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
	s.test = true
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

func drainFanoutQueue(s *Server) {
	//Run the queue to process the write
	go s.processFanoutQueue()
	for len(s.fanoutQueue) > 0 {
		time.Sleep(time.Second)
	}
	close(s.fanoutQueue)
}

func TestReadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)
	drainFanoutQueue(s)

	resp, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if resp.GetValue().GetValue()[0] != data[0] {
		t.Errorf("Bad return: %v", resp)
	}

}

func TestReadWriteComplexKey(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "me/myself/and/i", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)
	drainFanoutQueue(s)

	resp, err := s.Read(context.Background(), &pb.ReadRequest{Key: "me/myself/and/i"})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if resp.GetValue().GetValue()[0] != data[0] {
		t.Errorf("Bad return: %v", resp)
	}

	// Let's verify that the key has been written to the correct location
	_, err = os.Open(".testreadwrite/me/myself/and/i")
	if err != nil {
		t.Errorf("File did not load: %v", err)

		details, err := exec.Command("find", ".testreadwrite/").Output()
		log.Printf("Find output: %v -> %v", err, string(details))
	}

	// Validate that the internal list is empty
	files, err := ioutil.ReadDir(".testreadwrite/internal/towrite/")
	if err != nil {
		t.Errorf("Bad read of internal write dir")
	}

	if len(files) != 0 {
		t.Errorf("There's still files in the internal write dir: %v", files[0].Name())
	}
}

func TestInternalWriteRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.WriteInternal(context.Background(), &pb.WriteInternalRequest{Timestamp: time.Now().Unix(), Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)
	drainFanoutQueue(s)

	resp, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if resp.GetValue().GetValue()[0] != data[0] {
		t.Errorf("Bad return: %v", resp)
	}

}

func TestBadInternalWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badWrite = true

	data := []byte("magic")

	_, err := s.WriteInternal(context.Background(), &pb.WriteInternalRequest{Timestamp: time.Now().Unix(), Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
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

func TestBadReadOnRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badRead = true

	_, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestVirginRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.noConsensus = true

	val, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing", Consensus: int32(0)})
	if err == nil {
		t.Fatalf("Bad rad: %v", val)
	}
}

func TestVirginReadWithConsensus(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	val, err := s.Read(ctx, &pb.ReadRequest{Key: "testing", Consensus: int32(0)})
	if err != nil {
		t.Fatalf("Bad rad: %v", err)
	}

	if val.GetValue().GetValue()[0] != 'm' {
		t.Errorf("Bad read: %v", val)
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

func TestBadFanoutWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badFanoutWrite = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", err)
	}
}

func TestBadFanoutRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.badFanoutRead = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	drainQueue(s)
	drainFanoutQueue(s)

	if s.badQueueProcess != 1 {
		t.Errorf("Queue was processed fine")
	}
}

func TestBadFanoutSend(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")
	s.failFanout = true

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}

	// This will continually fail
	go s.processFanoutQueue()

	time.Sleep(time.Second * 2)

	if len(s.fanoutQueue) == 0 {
		t.Errorf("Somehow the fanout queue has been processed?")
	}

	// Trip fail the queue
	close(s.fanoutQueue)
}
