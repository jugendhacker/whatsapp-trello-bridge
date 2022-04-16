package server

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
)

func StartServer(next chan bool) {
	server := gin.New()
	server.Use(gin.Recovery())

	// required for creating webhook
	server.HEAD("/callback", func(c *gin.Context) {
		c.Status(200)
	})

	server.POST("/callback", func(c *gin.Context) {
		bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println(err)
			c.Status(500)
			return
		}
		if !VerifySignature(&c.Request.Header, bytes) {
			fmt.Println("Invalid signature")
			c.Status(403)
			return
		}
		c.Status(200)
		Handle(bytes)
	})
	go func() {
		time.Sleep(time.Second)
		next <- true
	}()
	fmt.Println("Starting webhook server")
	server.Run()
}
