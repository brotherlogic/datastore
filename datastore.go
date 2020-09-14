package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

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
	basepath   string
	friends    []string
	badWrite   bool
	badFanout  bool
	writeQueue chan string
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:   &goserver.GoServer{},
		basepath:   "/media/keystore/datastore/",
		writeQueue: make(chan string, 100),
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

	fmt.Printf("%v", server.Serve())
}
