package main

import (
	"context"
	"os"
	"testing"

	pb "github.com/brotherlogic/datastore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

func InitTest(remove bool, dir string) *Server {
	s := Init()
	s.SkipLog = true
	s.basepath = dir
	if remove {
		os.RemoveAll(dir)
	}
	return s
}

func TestBadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	s.testingfails = true
	grr, err := s.Write(context.Background(), &pb.WriteRequest{Key: "/testing", Value: &google_protobuf.Any{Value: data}})
	if err == nil {
		t.Fatalf("Bad write: %v", grr)
	}
}

func TestWriteBadRead(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "/testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}
	s.testingfails = true
	_, err = s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err == nil {
		t.Fatalf("Bad read: %v", err)
	}
}

func TestReadWrite(t *testing.T) {
	s := InitTest(true, ".testreadwrite/")

	data := []byte("magic")

	_, err := s.Write(context.Background(), &pb.WriteRequest{Key: "/testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad write: %v", err)
	}
	grr, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	data2 := grr.GetValue().GetValue()
	if len(data) != len(data2) {
		t.Errorf("Bytes came back wrong: %v vs %v", data, data2)
	}

	for i := range data {
		if data[i] != data2[i] {
			t.Errorf("Bad match %v %v", data, data2)
		}
	}

	_, err = s.Write(context.Background(), &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		t.Fatalf("Bad second write: %v", err)
	}

	grr, err = s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	data2 = grr.GetValue().GetValue()
	if len(data) != len(data2) {
		t.Errorf("Bytes came back wrong: %v vs %v", data, data2)
	}

	for i := range data {
		if data[i] != data2[i] {
			t.Errorf("Bad match %v %v", data, data2)
		}
	}
}

func TestBadRead(t *testing.T) {
	s := InitTest(true, ".testbadread/")

	grr, err := s.Read(context.Background(), &pb.ReadRequest{Key: "testing"})
	if err == nil {
		t.Fatalf("Not Bad read: %v", grr)
	}

}
