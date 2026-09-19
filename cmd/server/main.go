package main

import (
	"log"
	"os"

	"github.com/Lunidra/Ether-Backend/internal/websocket"
)

func main() {
	addr := os.Getenv("ETHER_BACKEND_ADDR")

	if addr == "" {
		addr = "0.0.0.0:2832"
	}
	if os.Getenv("ETHER_LINK_SECRET") == "" {
		log.Fatal("ETHER_LINK_SECRET is not configured")
	}

	server := websocket.NewServer(addr)

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
