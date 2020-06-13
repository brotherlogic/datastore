package main

import (
	"fmt"
	"time"

	pb "github.com/brotherlogic/datastore/proto"
	"github.com/brotherlogic/goserver/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//WriteQueueEntry holds an entry for fanout
type WriteQueueEntry struct {
	server       string
	writeRequest *pb.WriteRequest
	attempts     int
	ack          chan<- bool
	key          string
}

var (
	writeQueue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "write_queue_size",
		Help: "The size of the write queue",
	})
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Server) fanout(req *pb.WriteRequest, key string, count int) {
	rCount := 0
	ackChan := make(chan bool)
	for _, server := range s.friends {
		writeQueue.Set(float64(len(s.fanoutQueue)))
		s.fanoutQueue <- &WriteQueueEntry{server: server, writeRequest: req, key: key, ack: ackChan}
	}

	for rCount < min(count, len(s.friends)) {
		_, ok := <-ackChan
		if !ok {
			break
		}
		rCount++
	}

	close(ackChan)
	return
}

func (s *Server) handleFanout() {
	for {
		x, ok := <-s.fanoutQueue
		if !ok {
			break
		}

		ctx, cancel := utils.ManualContext("datastore-fanout", "datastore", time.Minute, false)
		defer cancel()
		conn, err := s.FDial(x.server)
		if err != nil {
			s.Log(fmt.Sprintf("Error on dial: %v", err))
			x.attempts++
			s.fanoutQueue <- x
		}
		defer conn.Close()

		client := pb.NewDatastoreServiceClient(conn)

		// Ensure we don't fanout fanned out writes
		x.writeRequest.FanoutMinimum = -1
		_, err = client.Write(ctx, x.writeRequest)
		if err != nil {
			s.Log(fmt.Sprintf("Error on write :%v", err))
			x.attempts++
			s.fanoutQueue <- x
		} else {
			x.ack <- true
		}
	}
}
