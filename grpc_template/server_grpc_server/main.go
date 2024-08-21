package main

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/Serpent1075/proto_example/pb/serverrpc"
	"google.golang.org/grpc"
)

type serverserv struct {
	pb.UnimplementedServerServiceServer
}

func (s *serverserv) ServerStreamingCall(in *pb.StreamingRequest, stream grpc.ServerStreamingServer[pb.StreamingResponse]) error {
	log.Printf("Received: %v", in.GetMessage())
	for i := 0; i < 5; i++ {
		if err := stream.Send(&pb.StreamingResponse{Message: fmt.Sprintf("Server message %d", i)}); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50052")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterServerServiceServer(s, &serverserv{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
