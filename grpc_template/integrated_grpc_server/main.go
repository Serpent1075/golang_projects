package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/Serpent1075/proto_example/pb/integratedrpc"
	"google.golang.org/grpc"
)

type unaryserv struct {
	pb.UnimplementedIntUnaryServiceServer
}

type serverserv struct {
	pb.UnimplementedIntServerServiceServer
}
type clientserv struct {
	pb.UnimplementedIntClientServiceServer
}

type bidirectionalserv struct {
	pb.UnimplementedIntBidirectionalServiceServer
}

func (s *unaryserv) UnaryCall(ctx context.Context, in *pb.UnaryRequest) (*pb.UnaryResponse, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &pb.UnaryResponse{Message: "Hello " + in.GetMessage()}, nil
}

func (s *clientserv) ClientStreamingCall(stream pb.IntClientService_ClientStreamingCallServer) error {
	var messages []string
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.StreamingResponse{Message: fmt.Sprintf("Received messages: %v", messages)})

		}
		if err != nil {
			return err
		}
		log.Printf("Received: %v", in.GetMessage())
		messages = append(messages, in.GetMessage())
	}
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

func (s *bidirectionalserv) BidirectionalStreamingCall(stream pb.IntBidirectionalService_BidirectionalStreamingCallServer) error {
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
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterIntUnaryServiceServer(s, &unaryserv{})
	pb.RegisterIntServerServiceServer(s, &serverserv{})
	pb.RegisterIntClientServiceServer(s, &clientserv{})
	pb.RegisterIntBidirectionalServiceServer(s, &bidirectionalserv{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
