package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	// "github.com/gallactic/gallactic/crypto"
	grpcode "github.com/gallactic/gallactic/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = "localhost:10903"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	fmt.Println("gprc is connected to", conn)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	c := grpcode.NewAccountsClient(conn)
	fmt.Println("gprc is registered", c)
	defer conn.Close()
	// Contact the server and print out its response.
	//address, _ := crypto.AddressFromString("ac9E2cyNA5UfB8pUpqzEz4QCcBpp8sxnEaN")
	var deadlineMs = flag.Int("deadline_ms", 15000*1000, "Default deadline in milliseconds.")
	fmt.Println("Timeout in Seconds : ", deadlineMs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*deadlineMs)*time.Millisecond)
	defer cancel()
	//fmt.Println("Address : ", address)
	f, ferr := c.GetValidators(ctx, &grpcode.Empty{})

	if ferr != nil {
		log.Fatalf("could not greet: %v", ferr)
	}
	log.Printf("Greeting: %s", f)
}
