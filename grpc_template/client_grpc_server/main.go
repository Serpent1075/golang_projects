package main

import (
	"fmt"
	"io"
	"log"
	"net"

	pb "github.com/Serpent1075/proto_example/pb/clientrpc"
	"google.golang.org/grpc"
)

type clientserv struct {
	pb.UnimplementedClientServiceServer
}

func (s *clientserv) ClientStreamingCall(stream pb.ClientService_ClientStreamingCallServer) error {
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

func main() {
	lis, err := net.Listen("tcp", "localhost:50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterClientServiceServer(s, &clientserv{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
