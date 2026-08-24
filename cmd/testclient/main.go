package main

import (
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(
		"ws://127.0.0.1:2832/ws",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	log.Println("Connected to Ether Backend")

	err = conn.WriteMessage(
		websocket.TextMessage,
		[]byte(`{
  "version": 1,
  "type": "auth_identify",
  "payload": {
    "uuid": "test-uuid",
    "username": "TestPlayer"
  }
}`),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Test packet sent")
}