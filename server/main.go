package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/adlio/trello"
	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow"
)

func StartServer(trelloClient *trello.Client, whatsAppClient *whatsmeow.Client) {
	server := gin.Default()

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
			c.Status(403)
			return
		}

		var request WebhookRequest
		err2 := json.Unmarshal(bytes, &request)

		if err2 != nil {
			fmt.Println(err2)
			c.Status(500)
			return
		}

		c.Status(200)
		Handle(trelloClient, whatsAppClient, request)
	})

	server.Run()
}
