package main

import (
	"fmt"
	"log"
	"os"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/datastore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	conn, err := grpc.Dial("discovery:///datastore", grpc.WithInsecure(), grpc.WithBalancerName("my_pick_first"))
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewDatastoreServiceClient(conn)
	ctx, cancel := utils.BuildContext("recordader-cli", "recordadder")
	defer cancel()

	switch os.Args[1] {
	case "write":
		res, err := client.Write(ctx, &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: []byte(os.Args[2])}})
		fmt.Printf("%v -> %v\n", res, err)
	}

}
