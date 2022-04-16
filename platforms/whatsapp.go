package platforms

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"time"

	"github.com/adlio/trello"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	store "github.com/drdeee/techsupport/store"
)

var WhatsAppClient *whatsmeow.Client

func InitWhatsAppClient() whatsmeow.Client {
	fmt.Println("Initializing WhatsApp client")
	dbLog := waLog.Stdout("Database", "WARN", true)
	container, err := sqlstore.New("sqlite3", "file:"+os.Getenv("WHATSAPP_DATABASE_FILE")+"?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("WhatsApp Client", "WARN", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}
	WhatsAppClient = client
	return *client
}

func eventHandler(event interface{}) {
	switch evt := event.(type) {
	case *events.Message:
		WhatsAppClient.MarkRead([]string{evt.Info.ID}, time.Now(), evt.Info.Chat, evt.Info.Sender)
		hasAttachment, attachmentFile, attachmentName, err := getAttachment(evt)
		if err != nil {
			if err.Error() != "type unsupported" {
				fmt.Println(err)
				SendText(*evt, "Der Anhang deiner Nachricht konnte nicht heruntergeladen werden :(")
			}
			return
		}
		state := store.Requests.GetState(evt.Info.Chat.String())
		if state == "" {
			var card = &trello.Card{
				Name:    getUsername(evt),
				Desc:    evt.Message.GetConversation(),
				IDBoard: os.Getenv("TRELLO_BOARD_ID"),
				IDList:  LIST_ID_NEW}
			err := TrelloClient.CreateCard(card)
			if err == nil {
				err = SetTrelloCustomFieldValue(card.ID, evt.Info.Sender.ToNonAD().String())
				if err == nil && hasAttachment {
					err = UploadTrelloAttachment(card.ID, attachmentFile, attachmentName)
				}
			}
			if err != nil {
				fmt.Println("Error creating card:", err)
				SendText(*evt, "Deine Anfrage konnte nicht weitergeleitet werden :( Bitte versuche es später nochmal erneut.")
			} else {
				store.Requests.SetState(evt.Info.Chat.ToNonAD().String(), card.ID)
				SendText(*evt, "Deine Anfrage wurde erfolgreich weitergeleitet. Wir kümmern uns so schnell wie möglich darum.")
			}
		} else {
			card, err := TrelloClient.GetCard(state)
			if err != nil {
				fmt.Println("Error adding comment to card:", err)
				SendText(*evt, "Deine Nachricht konnte nicht weitergeleitet werden :( Bitte versuche es später nochmal erneut.")
			} else {
				msg := "**[USER]** " + evt.Message.GetConversation()
				if hasAttachment {
					msg += "\n\n*(Neuer Anhang)* "
				}
				_, err := card.AddComment(msg)
				if err == nil && hasAttachment {
					err = UploadTrelloAttachment(card.ID, attachmentFile, attachmentName)
				}
				if err != nil {
					fmt.Println("Error adding comment to card:", err)
					SendText(*evt, "Deine Nachricht konnte nicht weitergeleitet werden :( Bitte versuche es später nochmal erneut.")
				} else {
					SendText(*evt, "Deine Nachricht wurde deiner Anfrage hinzugefügt.")
				}
			}
		}
	}
}

func getUsername(evt *events.Message) string {
	number := evt.Info.Sender.User
	contact, err := WhatsAppClient.Store.Contacts.GetContact(evt.Info.Sender)
	if err != nil || !contact.Found {
		if evt.Info.PushName != "" {
			return evt.Info.PushName + " (" + number + ")"
		}
		return number
	} else {
		if contact.BusinessName != "" {
			return contact.BusinessName + " (" + number + ")"
		} else if contact.FullName != "" {
			return contact.FullName + " (" + number + ")"
		} else {
			return evt.Info.Sender.User
		}
	}
}

func getAttachment(evt *events.Message) (bool, string, string, error) {
	var msg whatsmeow.DownloadableMessage
	var originalFileName string
	if evt.Message.GetVideoMessage() != nil {
		ext, err := getExtensionFromMimeType(evt.Message.GetVideoMessage().GetMimetype())
		if err != nil {
			return false, "", "", err
		}
		originalFileName = "video" + ext
		msg = evt.Message.GetVideoMessage()
	} else if evt.Message.GetAudioMessage() != nil {
		ext, err := getExtensionFromMimeType(evt.Message.GetAudioMessage().GetMimetype())
		if err != nil {
			return false, "", "", err
		}
		originalFileName = "audio" + ext
		msg = evt.Message.GetAudioMessage()
	} else if evt.Message.GetDocumentMessage() != nil {
		ext, err := getExtensionFromMimeType(evt.Message.GetDocumentMessage().GetMimetype())
		if err != nil {
			return false, "", "", err
		}
		originalFileName = evt.Message.GetDocumentMessage().GetFileName()
		if originalFileName == "" {
			originalFileName = "document" + ext
		}
		msg = evt.Message.GetDocumentMessage()
	} else if evt.Message.GetImageMessage() != nil {
		ext, err := getExtensionFromMimeType(evt.Message.GetImageMessage().GetMimetype())
		if err != nil {
			return false, "", "", err
		}
		originalFileName = "image" + ext
		msg = evt.Message.GetImageMessage()
	}
	if evt.Message.GetConversation() == "" && msg == nil {
		SendText(*evt, "Dieser Nachrichtentyp wird leider nicht unterstützt :(")
		return false, "", "", fmt.Errorf("type unsupported")
	}

	if msg != nil {
		file, err := WhatsAppClient.Download(msg)
		if err != nil {
			return false, "", "", err
		}
		fName, err := saveBytesToTempFile(file)
		if err != nil {
			return false, "", "", err
		}
		return true, fName, originalFileName, err
	} else {
		return false, "", "", nil
	}
}

func saveBytesToTempFile(data []byte) (string, error) {
	tmpfile, err := ioutil.TempFile("", "msg-media")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	if _, err := tmpfile.Write(data); err != nil {
		fmt.Println("d2")
		return "", err
	}
	return tmpfile.Name(), nil
}

func SendText(evt events.Message, err string) {
	WhatsAppClient.SendMessage(evt.Info.Chat, "", &waProto.Message{Conversation: proto.String(err)})
}

func SendTextWithJID(chatJID string, msg string) error {
	msgData := &waProto.Message{Conversation: proto.String(msg)}
	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return err
	}
	_, err = WhatsAppClient.SendMessage(jid.ToNonAD(), "", msgData)
	return err
}

func getExtensionFromMimeType(mimeType string) (string, error) {
	exts, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return "", err
	}
	if len(exts) > 0 {
		return exts[len(exts)-1], nil
	} else {
		return "", fmt.Errorf("no extension found")
	}
}
