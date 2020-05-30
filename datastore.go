package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/datastore/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	"github.com/brotherlogic/goserver/utils"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	basepath     string
	testingfails bool
	fanoutQueue  chan *WriteQueueEntry
	friends      []string
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

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.GoServer.KSclient = *keystoreclient.GetClient(server.DialMaster)
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("datastore", false, true)
	if err != nil {
		return
	}

	//Let our neighbours know about us
	servers, err := utils.ResolveAll("datastore")
	if err != nil {
		fmt.Printf("Error finding friends: %v\n", err)
		return
	}

	for _, serv := range servers {
		conn, err := server.DoDial(serv)
		if err != nil {
			fmt.Printf("Error dialling: %v", err)
			return
		}

		client := pb.NewDatastoreServiceClient(conn)
		ctx, cancel := utils.BuildContext("datastore-friend", "datastore")
		friend, err := client.Friend(ctx, &pb.FriendRequest{Friend: fmt.Sprintf("%v:%v", server.Registry.Identifier, server.Registry.Port)})
		if err == nil {
			server.friends = append(server.friends, friend.GetFriend())
		}
		cancel()
		conn.Close()
	}

	fmt.Printf("%v", server.Serve())
}
