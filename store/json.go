package store

import (
	"encoding/json"
	"fmt"
	"os"
)

type RequestStore map[string]string

func (store RequestStore) Init() {
	fmt.Println("Initializing store")
	bytes, err := os.ReadFile(os.Getenv("DATABASE_FILE"))
	if err != nil {
		fmt.Println("Could not read database file, creating new one")
		os.WriteFile(os.Getenv("DATABASE_FILE"), []byte("{}"), 0644)
	} else {
		err = json.Unmarshal(bytes, &store)
		if err != nil {
			panic("Could not parse database file")
		}
	}
}

func (store RequestStore) save() {
	bytes, err := json.Marshal(store)
	if err != nil {
		panic("Could not marshal requests")
	}
	os.WriteFile(os.Getenv("DATABASE_FILE"), bytes, 0644)
}

func (store RequestStore) SetState(whatsAppId, state string) {
	store[whatsAppId] = state
	store.save()
}

func (store RequestStore) GetState(whatsAppId string) (string, error) {
	if entry, ok := store[whatsAppId]; ok {
		return entry, nil
	}
	return "", fmt.Errorf("whatsAppId %v does not exist in store", whatsAppId)
}
