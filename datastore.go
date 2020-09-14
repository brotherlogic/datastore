package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/brotherlogic/goserver"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/datastore/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

var (
	//Friends - total number of friends
	Friends = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "datastore_friends",
		Help: "The number of friends we have",
	})
)

//Server main server type
type Server struct {
	*goserver.GoServer
	basepath        string
	friends         []string
	badWrite        bool
	badFanout       bool
	badMarshal      bool
	badBaseMarshal  bool
	badRead         bool
	badUnmarshal    bool
	badFanoutWrite  bool
	badFanoutRead   bool
	failFanout      bool
	writeQueue      chan string
	fanoutQueue     chan string
	cachedKey       map[string]bool
	badQueueProcess int
	test            bool
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:    &goserver.GoServer{},
		basepath:    "/media/keystore/datastore/",
		writeQueue:  make(chan string, 100),
		fanoutQueue: make(chan string, 100),
		cachedKey:   make(map[string]bool),
		friends:     make([]string, 0),
	}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterDatastoreServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "magic", Value: int64(13)},
	}
}

func (s *Server) fanout(ctx context.Context, req *pb.WriteInternalRequest) error {
	if s.failFanout {
		return fmt.Errorf("Failed")
	}

	if s.test {
		return nil
	}

	conn, err := s.FDialSpecificServer(ctx, "datastore", req.GetDestination())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewDatastoreInternalServiceClient(conn)
	_, err = client.WriteInternal(ctx, req)
	return err
}

func (s *Server) populateFriends(ctx context.Context) {
	if s.test {
		s.friends = []string{"test1"}
		return
	}
	vals, _ := s.FFind(ctx, "datastore")
	for _, server := range vals {
		elems := strings.Split(server, ":")
		found := false
		for _, blah := range s.friends {
			if blah == elems[0] {
				found = true
			}
		}

		if !found {
			s.friends = append(s.friends, elems[0])
		}
	}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("datastore", false, true)
	if err != nil {
		return
	}

	go server.processWriteQueue()

	fmt.Printf("%v", server.Serve())
}
