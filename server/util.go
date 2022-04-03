package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"os"
)

func CreateDigest(content string) string {
	// create hmac (sha1) of content and transform to base64
	h := hmac.New(sha1.New, []byte(os.Getenv("TRELLO_HASH_KEY")))

	// Write Data to it
	h.Write([]byte(content))

	// return encoded digest
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func VerifySignature(header *http.Header, body []byte) bool {
	return header.Get("X-Trello-Webhook") == CreateDigest(string(body)+os.Getenv("TRELLO_CALLBACK_URL"))
}
