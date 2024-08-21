package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	bidirectionpb "github.com/Serpent1075/proto_example/pb/bidirectionrpc"
	clientpb "github.com/Serpent1075/proto_example/pb/clientrpc"
	serverpb "github.com/Serpent1075/proto_example/pb/serverrpc"
	unarypb "github.com/Serpent1075/proto_example/pb/unaryrpc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	r := gin.Default()
	r.Use(JSONMiddleware())
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))
	v1 := r.Group("/v1")
	{
		v1.POST("/unary", unary_RPC)
		v1.POST("/server", server_stream_RPC)
		v1.POST("/client", client_stream_RPC)
		v1.POST("/bidirectional", bidirectional_stream_RPC)
	}
	r.Run(":3010")
}

func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0")
		c.Writer.Header().Set("Last-Modified", time.Now().String())
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "-1")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func init() {

}

type User struct {
	Name string `json:"name"`
}

func unary_RPC(c *gin.Context) {
	var u User
	var getnamedata error
	if getnamedata = c.ShouldBindJSON(&u); getnamedata != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	conn, err := grpc.NewClient("localhost:50051", grpc.EmptyDialOption{})
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	pbc := unarypb.NewUnaryServiceClient(conn)
	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()
	r, err := pbc.UnaryCall(ctx, &unarypb.UnaryRequest{Message: u.Name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.Message)

	c.JSON(200, gin.H{
		"message": "unary_RPC",
	})
}

func server_stream_RPC(c *gin.Context) {
	var u User
	var getnamedata error
	if getnamedata = c.ShouldBindJSON(&u); getnamedata != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	conn, err := grpc.NewClient("localhost:50052", grpc.EmptyDialOption{})
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	pbc := serverpb.NewServerServiceClient(conn)

	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()

	stream, err := pbc.ServerStreamingCall(ctx, &serverpb.StreamingRequest{Message: "Stream from client"})
	if err != nil {
		log.Fatalf("could not get stream: %v", err)
	}
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from stream: %s", response.GetMessage())
	}

	log.Println(u.Name)
	c.JSON(200, gin.H{
		"message": "server_stream_RPC",
	})
}

func client_stream_RPC(c *gin.Context) {
	var u User
	var getnamedata error
	if getnamedata = c.ShouldBindJSON(&u); getnamedata != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	conn, err := grpc.NewClient("localhost:50053", grpc.EmptyDialOption{})
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	pbc := clientpb.NewClientServiceClient(conn)

	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()

	stream, err := pbc.ClientStreamingCall(ctx)
	if err != nil {
		log.Fatalf("could not get stream: %v", err)
	}
	for i := 0; i < 5; i++ {
		if err := stream.Send(&clientpb.StreamingRequest{Message: fmt.Sprintf("Client message %d", i)}); err != nil {
			log.Fatalf("error while sending: %v", err)
		}
		time.Sleep(time.Second)
	}
	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving: %v", err)
	}
	log.Printf("Client streaming response: %s", reply.GetMessage())

	c.JSON(200, gin.H{
		"message": "client_stream_RPC",
	})
}

func bidirectional_stream_RPC(c *gin.Context) {
	var u User
	var getnamedata error
	if getnamedata = c.ShouldBindJSON(&u); getnamedata != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	log.Println(u.Name)

	conn, err := grpc.NewClient("localhost:50053", grpc.EmptyDialOption{})
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	pbc := bidirectionpb.NewBidirectionServiceClient(conn) // c 변수 정의

	ctx, cancel := context.WithTimeout(c, time.Second)
	defer cancel()

	stream, err := pbc.BidirectionStreamingCall(ctx)
	if err != nil {
		log.Fatalf("could not get stream: %v", err)
	}

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// 서버 스트림 종료
				return
			}
			if err != nil {
				log.Fatalf("Error receiving from server: %v", err)
			}
			log.Printf("Received from server: %s", in.GetMessage())
		}
	}()

	// 클라이언트에서 서버로 메시지 보내기
	for i := 0; i < 5; i++ {
		if err := stream.Send(&bidirectionpb.StreamingRequest{Message: fmt.Sprintf("Client message %d", i)}); err != nil {
			log.Fatalf("Error sending to server: %v", err)
		}
		time.Sleep(time.Second)
	}
	stream.CloseSend() // 클라이언트 스트림 종료

	// 서버 스트림이 완전히 닫힐 때까지 대기 (선택 사항)
	<-ctx.Done()

	c.JSON(200, gin.H{
		"message": "bidirectional_stream_RPC",
	})

}
