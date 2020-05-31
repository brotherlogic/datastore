package main

import (
	pb "github.com/brotherlogic/datastore/proto"
	"github.com/brotherlogic/goserver/utils"
)

//WriteQueueEntry holds an entry for fanout
type WriteQueueEntry struct {
	server       string
	writeRequest *pb.WriteRequest
	attempts     int
	ack          chan<- bool
}

func (s *Server) fanout(req *pb.WriteRequest) {
	for _, server := range s.friends {
		s.fanoutQueue <- &WriteQueueEntry{server: server, writeRequest: req}
	}
}

func (s *Server) handleFanout() {
	for {
		x, ok := <-s.fanoutQueue
		if !ok {
			break
		}

		conn, err := s.BaseDial(x.server)
		if err != nil {
			x.attempts++
			s.fanoutQueue <- x
		}
		defer conn.Close()

		client := pb.NewDatastoreServiceClient(conn)
		ctx, cancel := utils.BuildContext("datastore-fanout", "datastore")
		defer cancel()

		_, err = client.Write(ctx, x.writeRequest)
		if err != nil {
			x.attempts++
			s.fanoutQueue <- x
		} else {
			x.ack <- true
		}
	}
}
