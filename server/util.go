package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"os"
)

func (server *Server) createDigest(content string) string {
	// create hmac (sha1) of content and transform to base64
	h := hmac.New(sha1.New, []byte(os.Getenv("TRELLO_HASH_KEY")))

	// Write Data to it
	h.Write([]byte(content))

	// return encoded digest
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (server *Server) verifySignature(header *http.Header, body []byte) bool {
	return header.Get("X-Trello-Webhook") == server.createDigest(string(body)+server.trello.GetURL("/callback"))
}

func (server *Server) isOnline() bool {
	// do get request
	resp, err := http.Get(server.trello.GetURL("/status"))
	if err != nil {
		return false
	}

	if resp.StatusCode != 200 {
		return false
	}
	return true
}
