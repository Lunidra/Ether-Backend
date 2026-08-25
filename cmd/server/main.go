package main

import (
	"log"
	"os"

	"github.com/Lunidra/Ether-Backend/internal/websocket"
)

func main() {
	addr := os.Getenv("ETHER_BACKEND_ADDR")

	if addr == "" {
		addr = "127.0.0.1:2832"
	}

	//	server := websocket.NewServer(addr)
	server := websocket.NewDevelopmentServer(":2832")

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
