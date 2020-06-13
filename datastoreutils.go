package main

import (
	"fmt"
	"sync"
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
	ackChan := make(chan bool, 20)
	for _, server := range s.friends {
		writeQueue.Set(float64(len(s.fanoutQueue)))
		s.fanoutQueue <- &WriteQueueEntry{server: server, writeRequest: req, key: key, ack: ackChan}
	}

	var wg sync.WaitGroup
	wg.Add(min(count, len(s.friends)))
	for rCount < min(count, len(s.friends)) {
		_, ok := <-ackChan
		s.Log(fmt.Sprintf("Got ACK: %v and %v, %v", rCount, count, len(s.friends)))
		if !ok {
			break
		}
		rCount++
		wg.Done()
	}

	// Background wait for the remaining channels to ack before closing
	go func() {
		wg.Wait()
		s.Log(fmt.Sprintf("Closing the channel with %v and %v", rCount, wg))
		close(ackChan)
	}()

	return
}

func (s *Server) handleFanout() {
	for {
		x, ok := <-s.fanoutQueue
		s.Log(fmt.Sprintf("Read from queue: %v %v", x, ok))
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
