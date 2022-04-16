package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/Jeffail/gabs"
	"github.com/adlio/trello"
)

var TrelloClient *trello.Client
var customFieldId string

var LIST_ID_NEW, LIST_ID_DONE string

func InitTrelloClient() *trello.Client {
	fmt.Println("Initializing trello client")

	// setup variables
	LIST_ID_NEW, LIST_ID_DONE = os.Getenv("TRELLO_LIST_NEW_ID"), os.Getenv("TRELLO_LIST_DONE_ID")

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

	checkLists(board)
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
				customFieldId = field.ID
				return
			} else {
				panic("Custom field exists but is not of type text")
			}
		}
	}

	// create custom field
	err = createCustomField()
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

func checkLists(board *trello.Board) {
	lists, err := board.GetLists()
	if err != nil {
		panic("Could not load list of board")
	}

	var new bool
	var done bool

	for _, list := range lists {
		switch list.ID {
		case LIST_ID_NEW:
			new = true
		case LIST_ID_DONE:
			done = true
		}
	}

	if !new || !done {
		data, _ := fmt.Printf("Some list not found (new: %t, done: %t)\n", new, done)

		// print all names and ids of lists
		for _, list := range lists {
			fmt.Printf("%s: %s\n", list.ID, list.Name)
		}
		panic(data)
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

func createCustomField() error {
	reqBody, err := json.Marshal(map[string]string{
		"idModel":   os.Getenv("TRELLO_BOARD_ID"),
		"modelType": "board",
		"name":      os.Getenv("TRELLO_CUSTOM_FIELD_NAME"),
		"type":      "text",
		"pos":       "top",
	})
	if err != nil {
		return err
	}
	// make http post
	var path = "https://api.trello.com/1/customFields?key=" + os.Getenv("TRELLO_KEY") + "&token=" + os.Getenv("TRELLO_TOKEN") + "&pos=top"
	res, err := http.Post(path, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("error creating custom field")
	}

	body, err := gabs.ParseJSONBuffer(res.Body)
	if err != nil {
		return err
	}

	customFieldId = body.Path("id").Data().(string)
	return nil
}

func GetTrelloCustomFieldValue(cardId string) (string, error) {
	//make http get
	var path = "https://api.trello.com/1/cards/" + cardId + "/customFieldItems?key=" + os.Getenv("TRELLO_KEY") + "&token=" + os.Getenv("TRELLO_TOKEN")
	res, err := http.Get(path)
	if err != nil {
		return "", fmt.Errorf("Error getting trello custom field value: " + err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "nil", err
	}

	var arr []map[string]interface{}
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return "", err
	}

	for _, field := range arr {
		if field["idCustomField"].(string) == customFieldId {
			return field["value"].(map[string]interface{})["text"].(string), nil
		}
	}
	return "", fmt.Errorf("could not find value of custom field")
}

func SetTrelloCustomFieldValue(cardId string, value string) error {
	//make http put
	var path = "https://api.trello.com/1/card/" + cardId + "/customField/" + customFieldId + "/item?key=" + os.Getenv("TRELLO_KEY") + "&token=" + os.Getenv("TRELLO_TOKEN")
	req, err := http.NewRequest("PUT", path, bytes.NewBuffer([]byte(`{"value":{"text":"`+value+`"}}`)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err1 := http.DefaultClient.Do(req)
	if err1 != nil {
		return err1
	} else if res.StatusCode != 200 {
		return fmt.Errorf("Invalid status code: " + res.Status)
	}
	return nil
}

func UploadTrelloAttachment(cardId, filePath, fileName string) error {
	var path = "https://api.trello.com/1/cards/" + cardId + "/attachments?key=" + os.Getenv("TRELLO_KEY") + "&token=" + os.Getenv("TRELLO_TOKEN")
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileWriter, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	w.Close()

	os.Remove(filePath)
	req, err := http.NewRequest("POST", path, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Invalid status code: " + res.Status)
	} else {
		return nil
	}
}

/*
func deleteTWebhook(webhookId string) error {
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
