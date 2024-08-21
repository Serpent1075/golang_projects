package main

import (
	"io"
	"log"
	"net"

	pb "github.com/Serpent1075/proto_example/pb/bidirectionrpc"
	"google.golang.org/grpc"
)

type bidirectionalserv struct {
	pb.UnimplementedBidirectionServiceServer
}

func (s *bidirectionalserv) BidirectionStreamingCall(stream pb.BidirectionService_BidirectionStreamingCallServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("Received from client: %s", in.GetMessage())
		if err := stream.Send(&pb.StreamingResponse{Message: "Hello from the server!"}); err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBidirectionServiceServer(s, &bidirectionalserv{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
