package platforms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/adlio/trello"
)

var TrelloClient *trello.Client

func InitTrelloClient() *trello.Client {
	fmt.Println("Initializing trello client")

	// create trello client
	client := trello.NewClient(os.Getenv("TRELLO_KEY"), os.Getenv("TRELLO_TOKEN"))

	// get trello board
	board, err := client.GetBoard(os.Getenv("TRELLO_BOARD_ID"))
	if err != nil {
		// trello board not found, stop execution
		panic("Invalid or missing environment variables: " + err.Error())
	}
	fmt.Println("Trello board found: " + board.Name)

	checkCustomFields(board)

	if os.Getenv("TRELLO_HASH_KEY") == "" {
		panic("Trello hash key not set")
	}
	TrelloClient = client
	return client
}

func checkCustomFields(board *trello.Board) {
	fields, err := board.GetCustomFields()
	if err != nil {
		panic("Error getting custom fields: " + err.Error())
	}

	// iterate over custom fields
	for _, field := range fields {
		if field.Name == os.Getenv("TRELLO_CUSTOM_FIELD_NAME") {
			if field.Type == "text" {
				return
			} else {
				panic("Custom field exists but is not of type text")
			}
		}
	}

	// create custom field
	err = createCustomField(board.ID)
	if err != nil {
		panic("Could not create custom field: " + err.Error())
	}
}

func CheckTrelloWebhooks() {

	var webhooks []trello.Webhook

	webhooks, err := getWebhooks()
	if err != nil {
		panic("Error getting trello webhooks: " + err.Error())
	}

	var exists bool = false

	// check if webhook has correct callback url
	for _, webhook := range webhooks {
		if webhook.CallbackURL == os.Getenv("TRELLO_CALLBACK_URL") {
			exists = true
		}
	}

	if !exists {
		err := createWebhook(os.Getenv("TRELLO_BOARD_ID"))
		if err != nil {
			panic("Error creating trello webhook: " + err.Error())
		}
	}
}

func createWebhook(boardId string) error {
	// make http post
	var path = "https://api.trello.com/1/tokens/" + os.Getenv("TRELLO_TOKEN") + "/webhooks?key=" + os.Getenv("TRELLO_KEY") + "&callbackURL=" + os.Getenv("TRELLO_CALLBACK_URL") + "&idModel=" + boardId
	res, err := http.Post(path, "application/json", nil)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("error creating webhook")
	}

	return nil
}

func getWebhooks() ([]trello.Webhook, error) {
	resp, err := http.Get("https://api.trello.com/1/tokens/" + os.Getenv("TRELLO_TOKEN") + "/webhooks?key=" + os.Getenv("TRELLO_KEY"))

	var webhooks []trello.Webhook
	decoder := json.NewDecoder(resp.Body)
	err1 := decoder.Decode(&webhooks)
	if err1 != nil || err != nil {
		return nil, err1
	} else {
		return webhooks, nil
	}
}

func createCustomField(boardId string) error {
	// make http post
	var path = "https://api.trello.com/1/customFields?key=" + os.Getenv("TRELLO_KEY") + "&name=" + os.Getenv("TRELLO_CUSTOM_FIELD_NAME") + "&type=text&modelType=board&modelId=" + boardId + "&token=" + os.Getenv("TRELLO_TOKEN") + "&pos=top"
	res, err := http.Post(path, "application/json", nil)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		fmt.Println("Custom field created")
		return fmt.Errorf("error creating custom field")
	}

	return nil
}

/* func deleteTWebhook(webhookId string) error {
	var path = "https://api.trello.com/1/tokens/" + os.Getenv("TRELLO_TOKEN") + "/webhooks/" + webhookId + "?key=" + os.Getenv("TRELLO_KEY")
	// make http delete
	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return err
	}
	res, err1 := http.DefaultClient.Do(req)
	if err1 != nil {
		return err1
	}
	fmt.Print(res.Status)
	return nil
}
*/
