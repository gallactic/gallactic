package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gallactic/gallactic/crypto"
	grpcode "github.com/gallactic/gallactic/grpc"
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
	c := grpcode.NewAccountsClient(conn)
	fmt.Println("gprc is registered", c)
	defer conn.Close()
	// Contact the server and print out its response.
	address, _ := crypto.AddressFromString("acB1fzw8un1P7T2rvdwtFaGBNiju6nHbVje")
	valaddress, _ := crypto.AddressFromString("vaBj9Lzb79w5NttG6VDctoKGZJWyZYJ2uoz")
	var deadlineMs = flag.Int("deadline_ms", 15000*1000, "Default deadline in milliseconds.")
	fmt.Println("Timeout in Seconds : ", deadlineMs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*deadlineMs)*time.Millisecond)
	defer cancel()
	fmt.Println("Address : ", address)
	//f, ferr := c.GetAccount(ctx, &grpcode.AddressRequest{Address: address})
	f, ferr := c.GetValidator(ctx, &grpcode.AddressRequest{Address: valaddress})
	if ferr != nil {
		log.Fatalf("could not greet: %v", ferr)
	}
	fmt.Printf("Greeting: %#v", f)

}
