package main

import (
	"fmt"
	"log"

	pb "github.com/tjfoc/tjfoc/protos/peer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "10.1.3.150:8000"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewPeerClient(conn)

	r, err := c.BlockchainGetHeight(context.Background(), &pb.BlockchainBool{true})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Printf("+++++++++++r = %v+++++++++++++\n", r)
}
