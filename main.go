package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adlio/trello"
	"github.com/drdeee/techsupport/server"
	"github.com/drdeee/techsupport/util"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// create trello client
	trelloClient := trello.NewClient(os.Getenv("TRELLO_KEY"), os.Getenv("TRELLO_TOKEN"))
	fmt.Println("Trello client created")

	// get trello board
	board, err := trelloClient.GetBoard(os.Getenv("TRELLO_BOARD_ID"))
	if err != nil {
		// trello board not found, stop execution
		fmt.Printf("Invalid or missing environment variables: %s", err.Error())
		return
	}
	fmt.Println("Trello board found: " + board.Name)

	if os.Getenv("TRELLO_HASH_KEY") == "" {
		fmt.Println("Trello hash key not set")
		return
	}

	// create whatsapp client
	whatsappClient := util.InitWhatsAppClient()
	go server.StartServer(trelloClient, &whatsappClient)

	fmt.Println("Waiting a second")
	time.Sleep(time.Second)

	// get all webhooks
	err1 := util.CheckWebhooks()
	if err1 == nil {
		fmt.Println("Webhook exists")
	} else {
		fmt.Println("Webhooks could not be created")
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	whatsappClient.Disconnect()
}
