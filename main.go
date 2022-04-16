package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/drdeee/techsupport/platforms"
	"github.com/drdeee/techsupport/server"
	store "github.com/drdeee/techsupport/store"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	store.Requests.Init()

	platforms.InitWhatsAppClient()
	platforms.InitTrelloClient()

	next := make(chan bool, 1)
	go server.StartServer(next)
	<-next

	// check webhooks
	platforms.CheckTrelloWebhooks()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	platforms.WhatsAppClient.Disconnect()
}
