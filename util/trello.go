package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/adlio/trello"
)

// source: https://github.com/adlio/trello/blob/02f26cf22237287b3c7ad18890d9adb481438723/webhook.go#L88
func CheckWebhooks() error {

	var webhooks []trello.Webhook

	webhooks, err := GetWebhooks()
	if err != nil {
		return err
	}

	var exists bool = false

	// check if webhook has correct callback url
	for _, webhook := range webhooks {
		if webhook.CallbackURL == os.Getenv("TRELLO_CALLBACK_URL") {
			exists = true
		}
	}

	if !exists {
		err := CreateWebhook(os.Getenv("TRELLO_BOARD_ID"))
		return err
	}

	return nil

}

func CreateWebhook(boardId string) error {
	// make http post
	var path = "https://api.trello.com/1/tokens/" + os.Getenv("TRELLO_TOKEN") + "/webhooks?key=" + os.Getenv("TRELLO_KEY") + "&callbackURL=" + os.Getenv("TRELLO_CALLBACK_URL") + "&idModel=" + boardId
	res, err := http.Post(path, "application/json", nil)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		return fmt.Errorf("error creating webhook")
	}

	fmt.Print(res.StatusCode)
	return nil
}

func GetWebhooks() ([]trello.Webhook, error) {
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

func DeleteWebhook(webhookId string) error {
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
