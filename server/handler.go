package server

import (
	"fmt"

	"github.com/Jeffail/gabs"
)

func Handle(data []byte) {
	request, err := gabs.ParseJSON(data)
	if err != nil {
		fmt.Printf("Error parsing webhook JSON: %s\n", err.Error())
		return
	}
	actionType := request.Path("action.type").Data().(string)
	switch actionType {
	case "updateCard":
		handleUpdateCard(request)
	case "commentCard":
		handleCommentCard(request)
	default:
		fmt.Printf("Unkhandled action type: %s\n", actionType)
	}

}

func handleUpdateCard(request *gabs.Container) {
	fmt.Println(request.EncodeJSON())
}

func handleCommentCard(request *gabs.Container) {
	fmt.Println(request.EncodeJSON())
}
