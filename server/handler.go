package server

import (
	"fmt"

	"github.com/adlio/trello"
	"go.mau.fi/whatsmeow"
)

func Handle(trello *trello.Client, whatsApp *whatsmeow.Client, request WebhookRequest) {
	fmt.Println("Handling webhook request")
}
