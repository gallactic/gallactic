package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gallactic/gallactic/crypto"
	pb "github.com/gallactic/gallactic/rpc/grpc/proto3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = "127.0.0.1:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	fmt.Println("gprc is connected to", conn)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := pb.NewBlockChainClient(conn)
	fmt.Println("gprc is registered", c)
	defer conn.Close()
	// Contact the server and print out its response.
	address, _ := crypto.AddressFromString("acB1fzw8un1P7T2rvdwtFaGBNiju6nHbVje")
	//valaddress, _ := crypto.AddressFromString("vaRHS4pqRDNAFaRjfLkgeAgDcnG115tT6R3")
	var deadlineMs = flag.Int("deadline_ms", 15000*1000, "Default deadline in milliseconds.")
	fmt.Println("Timeout in Seconds : ", deadlineMs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*deadlineMs)*time.Millisecond)
	defer cancel()
	fmt.Println("Address : ", address)
	f, ferr := c.GetAccounts(ctx, &pb.Empty{})
	//f, ferr := c.GetValidators(ctx, &pb.Empty{})
	if ferr != nil {
		log.Fatalf("could not greet: %v", ferr)
	}
	fmt.Printf("Greeting: %#v", f)

}
