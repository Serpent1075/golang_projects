package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	auth "github.com/whitewhale1075/urmy_rest_handler/auth"
)

func main() {
	a := InitAppHanlder()
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(JSONMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := r.Group("/v1")
	{
		v1.POST("/signup", a.signup)
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

func (a *AppHandler) signup(c *gin.Context) {
	data := auth.InsertUserProfile{
		DeviceOS: "test",
		LoginID:  "test1234",
		Useruid:  "test1234",
		Name:     "JohnoH",
	}

	a.authhandler.RegisterUser(&data)
	c.JSON(http.StatusOK, gin.H{})
}

func InitAppHanlder() *AppHandler {
	return &AppHandler{
		authhandler: auth.NewAuthHandler("testpath"),
	}
}
