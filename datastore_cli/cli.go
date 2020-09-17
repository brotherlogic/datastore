package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/datastore/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"

	//Needed to pull in gzip encoding init
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	all, err := utils.BaseResolveAll("datastore")
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	chosen := all[0]
	fmt.Printf("Writing to %v\n", chosen)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", chosen.Identifier, chosen.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewDatastoreServiceClient(conn)
	ctx, cancel := utils.ManualContext("recordader-cli", "recordadder", time.Minute, false)
	defer cancel()

	switch os.Args[1] {
	case "write":
		res, err := client.Write(ctx, &pb.WriteRequest{Key: "testing", Value: &google_protobuf.Any{Value: []byte(os.Args[2])}})
		fmt.Printf("%v -> %v\n", res, err)
	case "read":
		res, err := client.Read(ctx, &pb.ReadRequest{Key: "testing"})
		fmt.Printf("%v -> %v\n", res, err)
	}

}
