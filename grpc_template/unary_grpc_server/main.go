package main

import (
	"context"
	"log"
	"net"

	unarypb "github.com/Serpent1075/proto_example/pb/unaryrpc"
	"google.golang.org/grpc"
)

type server struct {
	unarypb.UnimplementedUnaryServiceServer
}

func (s *server) UnaryCall(ctx context.Context, in *unarypb.UnaryRequest) (*unarypb.UnaryResponse, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &unarypb.UnaryResponse{Message: "Hello " + in.GetMessage()}, nil
}

func main() {

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	unarypb.RegisterUnaryServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
