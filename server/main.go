package server

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/drdeee/whatsapp-trello-bridge/platforms"
	"github.com/drdeee/whatsapp-trello-bridge/store"
	"github.com/gin-gonic/gin"
)

type Server struct {
	trello   *platforms.TrelloClient
	whatsApp *platforms.WhatsAppClient
	store    *store.RequestStore
	server   *gin.Engine
}

var PublicUrl string

func (s *Server) Start(next chan bool, trelloClient *platforms.TrelloClient, whatsappClient *platforms.WhatsAppClient, store *store.RequestStore) {
	s.trello, s.whatsApp, s.store = trelloClient, whatsappClient, store

	s.server = gin.New()
	s.server.Use(gin.Recovery())

	// required for check availability
	s.server.GET("/status", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "online",
		})
	})

	// required for creating webhook
	s.server.HEAD("/callback", func(c *gin.Context) {
		c.Status(200)
	})

	s.server.POST("/callback", func(c *gin.Context) {
		bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println(err)
			c.Status(500)
			return
		}
		if !s.verifySignature(&c.Request.Header, bytes) {
			fmt.Println("Invalid signature")
			c.Status(403)
			return
		}
		c.Status(200)
		if s.whatsApp.IsReady() {
			s.handle(bytes)
		}
	})
	go func() {
		time.Sleep(time.Second)
		trys := 0

		for trys < 10 {
			trys++
			if s.isOnline() {
				next <- true
				return
			}
			fmt.Println("Could not connect to status endpoint, trying again in 10 seconds")
			time.Sleep(time.Second * 10)
		}
		panic("Server is not publicy available")
	}()
	fmt.Println("Starting webhook server")
	s.server.Run()
}
