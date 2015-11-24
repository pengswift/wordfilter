package main

import (
	"log"

	pb "github.com/pengswift/wordfilter/wordfilter"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address = "localhost:60051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewWordFilterServiceClient(conn)

	r, err := c.Filter(context.Background(), &pb.WordFilterRequest{Text: "我操你大爷，法轮大法好"})
	if err != nil {
		log.Fatalf("could not filter: %v", err)
	}
	log.Printf("text: %s", r.Text)
}
