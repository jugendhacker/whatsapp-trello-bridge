package server

import (
	"fmt"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/adlio/trello"
	"github.com/drdeee/techsupport/platforms"
	"github.com/drdeee/techsupport/store"
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
	case "updateCard":
		handleUpdateCard(request, card, field)
	case "commentCard":
		handleCommentCard(request, card, field)
	default:
		fmt.Printf("Unkhandled action type: %s\n", actionType)
	}

}

func handleUpdateCard(request *gabs.Container, card *trello.Card, field string) {
	if request.Path("action.data.listAfter").Data() != nil {
		if request.Path("action.data.listAfter.id").Data().(string) == platforms.LIST_ID_DONE {
			// card moved to done list, closing it
			store.Requests.SetState(field, "")
			err := platforms.SendTextWithJID(field, "Dein Ticket wurde geschlossen. Falls du der Meinung bist, dass dein Problem noch nicht gelöst wurde, kannst du gerne einfach eine weitere Nachricht schreiben :)")
			if err != nil {
				card.AddComment("**[BOT]** Die Quelle dieser Karte ist ungültig. Deine Nachricht konnte nicht weitergeleitet werden.")
			}
			card.AddComment("**[BOT]** Dieses Ticket wurde geschlossen.")
			platforms.SetTrelloCustomFieldValue(card.ID, "")
		}
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
