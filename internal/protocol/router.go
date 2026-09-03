package protocol

import (
	"fmt"
	"log"

	"github.com/Lunidra/Ether-Backend/internal/session"
)

const maxPacketsPerSecond = 30

type Handler func(
	client *session.Client,
	message Message,
) error

type Router struct {
	handlers map[string]Handler
}

func NewRouter() *Router {
	return &Router{
		handlers: make(
			map[string]Handler,
		),
	}
}

func (r *Router) Register(
	messageType string,
	handler Handler,
) {
	r.handlers[messageType] = handler
}

func (r *Router) Handle(
	client *session.Client,
	message Message,
) error {

	// Version 0 means this is the existing
	// Ether wire protocol, which is currently
	// what the real client uses.
	if message.Version != 0 &&
		message.Version != Version {

		return fmt.Errorf(
			"unsupported protocol version: %d",
			message.Version,
		)
	}

	if !client.AllowPacket(maxPacketsPerSecond) {
		return fmt.Errorf(
			"packet rate limit exceeded",
		)
	}

	if !client.Authenticated &&
		message.Type != "auth_hello" &&
		message.Type != "auth_verify" {

		return fmt.Errorf(
			"authentication required",
		)
	}

	if message.Type == "auth_verify" &&
		!client.Identified {
		return fmt.Errorf("client must identify before verification")
	}

	if client.Authenticated &&
		(message.Type == "auth_hello" ||
			message.Type == "auth_verify") {

		return fmt.Errorf(
			"client is already authenticated",
		)
	}

	handler, exists :=
		r.handlers[message.Type]

	if !exists {
		return fmt.Errorf(
			"unknown message type: %s",
			message.Type,
		)
	}

	return handler(
		client,
		message,
	)
}

func (r *Router) RegisterDebugHandlers() {

	r.Register(
		"test",
		func(
			client *session.Client,
			message Message,
		) error {

			log.Printf(
				"received test packet from client %s",
				client.ID,
			)

			return nil
		},
	)
}
