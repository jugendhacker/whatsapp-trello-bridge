package server

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/adlio/trello"
	"github.com/drdeee/whatsapp-trello-bridge/platforms"
	"github.com/drdeee/whatsapp-trello-bridge/store"
)

var actionWhiteList = []string{"updateCard", "commentCard"}

func Handle(data []byte) {
	request, err := gabs.ParseJSON(data)
	if err != nil {
		fmt.Printf("Error parsing webhook JSON: %s\n", err.Error())
		return
	}
	actionType := request.Path("action.type").Data().(string)

	var isWhitelisted bool
	for _, action := range actionWhiteList {
		if action == actionType {
			isWhitelisted = true
			break
		}
	}
	if !isWhitelisted {
		return
	}

	card, _ := platforms.TrelloClient.GetCard(request.Path("action.data.card.id").Data().(string))
	field, err := platforms.GetTrelloCustomFieldValue(card.ID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if field == "" {
		// card is not created by techsupport, ignoring it
		return
	}

	switch actionType {
	case "commentCard":
		handleCommentCard(request, card, field)
	case "updateCard":
		handleUpdateCard(request, card, field)
	}
}

func handleCommentCard(request *gabs.Container, card *trello.Card, field string) {
	text := request.Path("action.data.text").Data().(string)

	if strings.HasPrefix(text, "**[BOT]** ") || strings.HasPrefix(text, "**[USER]** ") {
		return
	}

	err := platforms.SendTextWithJID(field, text)
	if err != nil {
		card.AddComment("**[BOT]** Die Quelle dieser Karte ist ungültig. Deine Nachricht konnte nicht weitergeleitet werden.")
	}
}

func handleUpdateCard(request *gabs.Container, card *trello.Card, field string) {
	if request.Path("action.data.card").Data() != nil && request.Path("action.data.card.closed").Data().(bool) && request.Path("action.data.old").Data() != nil && !request.Path("action.data.old.closed").Data().(bool) {
		store.Requests.SetState(field, "")
		platforms.SetTrelloCustomFieldValue(card.ID, "")
		card.AddComment("**[BOT]** Diese Karte ist nun geschlossen, und wird keine weiteren Nachrichten mehr erhalten bzw. weiterleiten.")
	}
}
