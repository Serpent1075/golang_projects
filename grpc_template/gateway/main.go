package main

import (
	"log"
	"net/http"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
	log.Println(u.Name)
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
	c.JSON(200, gin.H{
		"message": "bidirectional_stream_RPC",
	})

}
