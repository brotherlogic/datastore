package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
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
	basepath      string
	testingfails  bool
	fanoutQueue   chan *WriteQueueEntry
	friends       []string
	validateCount int
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:    &goserver.GoServer{},
		basepath:    "/media/keystore/datastore/",
		fanoutQueue: make(chan *WriteQueueEntry),
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

func (s *Server) validate(ctx context.Context) error {
	//Let our neighbours know about us
	servers, err := utils.ResolveAll("datastore")
	if err != nil {
		fmt.Printf("Error finding friends: %v\n", err)
		return err
	}

	for _, serv := range servers {
		if serv.Identifier != s.Registry.Identifier {
			conn, err := s.FDial(fmt.Sprintf("%v:%v", serv.Identifier, serv.Port))
			if err != nil {
				fmt.Printf("Error dialling: %v", err)
				return err
			}

			client := pb.NewDatastoreServiceClient(conn)
			ctx, cancel := utils.BuildContext("datastore-friend", "datastore")
			friend, err := client.Friend(ctx, &pb.FriendRequest{Friend: append(s.friends, fmt.Sprintf("%v:%v", s.Registry.Identifier, s.Registry.Port))})
			if err == nil {
				for _, fr := range friend.GetFriend() {
					found := false
					for _, ffr := range s.friends {
						if ffr == fr {
							found = true
						}
					}
					if !found {
						s.friends = append(s.friends, fr)
					}

				}
				Friends.Set(float64(len(s.friends)))
			}
			cancel()
			conn.Close()
		}
	}

	s.validateCount++
	return nil
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

	server.RegisterRepeatingTaskNonMaster(server.validate, "validate", time.Minute)

	// Background run the fanout handler
	go server.handleFanout()

	fmt.Printf("%v", server.Serve())
}
