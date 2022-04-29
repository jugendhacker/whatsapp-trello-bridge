package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/Jeffail/gabs"
	"github.com/adlio/trello"
)

type TrelloClient struct {
	Client *trello.Client
	Board  *trello.Board
	Lists  struct {
		New, Done string
	}
	customFieldId string
	trelloAuth    string

	publicUrl url.URL
}

func (c *TrelloClient) GetURL(p string) string {
	u, _ := url.Parse(c.publicUrl.String())
	u.Path = path.Join(u.Path, p)
	return u.String()
}

func (c *TrelloClient) Init() {
	url, err := url.Parse(os.Getenv("PUBLIC_URL"))
	c.publicUrl = *url
	if err != nil {
		panic("Invalid public url: " + err.Error())
	}

	fmt.Println("Initializing trello client")

	// setup variables
	c.Lists.New, c.Lists.Done = os.Getenv("TRELLO_LIST_NEW_ID"), os.Getenv("TRELLO_LIST_DONE_ID")

	// create trello client
	c.Client = trello.NewClient(os.Getenv("TRELLO_KEY"), os.Getenv("TRELLO_TOKEN"))

	// get trello board
	c.Board, err = c.Client.GetBoard(os.Getenv("TRELLO_BOARD_ID"))
	if err != nil {
		// trello board not found, stop execution
		panic("Invalid or missing environment variables: " + err.Error())
	}
	fmt.Println("Trello board found: " + c.Board.Name)

	c.checkCustomFields()

	if os.Getenv("TRELLO_HASH_KEY") == "" {
		panic("Trello hash key not set")
	}

	c.checkLists()

	c.trelloAuth = "key=" + os.Getenv("TRELLO_KEY") + "&token=" + os.Getenv("TRELLO_TOKEN")
}

func (c *TrelloClient) checkCustomFields() {
	fields, err := c.Board.GetCustomFields()
	if err != nil {
		panic("Error getting custom fields: " + err.Error())
	}

	// iterate over custom fields
	for _, field := range fields {
		if field.Name == os.Getenv("TRELLO_CUSTOM_FIELD_NAME") {
			if field.Type == "text" {
				c.customFieldId = field.ID
				return
			} else {
				panic("Custom field exists but is not of type text")
			}
		}
	}

	// create custom field
	err = c.createCustomField()
	if err != nil {
		panic("Could not create custom field: " + err.Error())
	}
}

func (c *TrelloClient) CheckTrelloWebhooks() {

	var webhooks []trello.Webhook

	webhooks, err := c.getWebhooks()
	if err != nil {
		panic("Error getting trello webhooks: " + err.Error())
	}

	var exists bool = false

	// check if webhook has correct callback url
	for _, webhook := range webhooks {
		if webhook.CallbackURL == c.GetURL("/callback") {
			exists = true
		}
	}

	if !exists {
		err := c.createWebhook()
		if err != nil {
			panic("Error creating trello webhook: " + err.Error())
		}
	}
}

func (c *TrelloClient) checkLists() {
	lists, err := c.Board.GetLists()
	if err != nil {
		panic("Could not load list of board")
	}

	var new bool
	var done bool

	for _, list := range lists {
		switch list.ID {
		case c.Lists.New:
			new = true
		case c.Lists.Done:
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

func (c *TrelloClient) createWebhook() error {
	// make http post
	var path = "https://api.trello.com/1/tokens/" + os.Getenv("TRELLO_TOKEN") + "/webhooks?key=" + os.Getenv("TRELLO_KEY") + "&callbackURL=" + c.GetURL("/callback") + "&idModel=" + c.Board.ID
	res, err := http.Post(path, "application/json", nil)
	if err != nil {
		return err
	} else if res.StatusCode != 200 {
		// print body
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
		return fmt.Errorf("error creating webhook")
	}

	return nil
}

func (c *TrelloClient) getWebhooks() ([]trello.Webhook, error) {
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

func (c *TrelloClient) createCustomField() error {
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
	var path = "https://api.trello.com/1/customFields?" + c.trelloAuth + "&pos=top"
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

	c.customFieldId = body.Path("id").Data().(string)
	return nil
}

func (c *TrelloClient) GetTrelloCustomFieldValue(cardId string) (string, error) {
	//make http get
	var path = "https://api.trello.com/1/cards/" + cardId + "/customFieldItems?" + c.trelloAuth
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
		if field["idCustomField"].(string) == c.customFieldId {
			return field["value"].(map[string]interface{})["text"].(string), nil
		}
	}
	return "", fmt.Errorf("could not find value of custom field")
}

func (c *TrelloClient) SetTrelloCustomFieldValue(cardId string, value string) error {
	//make http put
	var path = "https://api.trello.com/1/card/" + cardId + "/customField/" + c.customFieldId + "/item?" + c.trelloAuth
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

func (c *TrelloClient) UploadTrelloAttachment(cardId, filePath, fileName string) error {
	var path = "https://api.trello.com/1/cards/" + cardId + "/attachments?" + c.trelloAuth
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
