package protocol

import (
	"fmt"
	"log"

	"github.com/Lunidra/Ether-Backend/internal/session"
)

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

	if !client.IsAuthenticated() &&
		message.Type != "auth_hello" &&
		message.Type != "auth_verify" {

		return fmt.Errorf(
			"authentication required",
		)
	}

	if message.Type == "auth_verify" &&
		!client.IsIdentified() {
		return fmt.Errorf("client must identify before verification")
	}

	if client.IsAuthenticated() &&
		(message.Type == "auth_hello" ||
			message.Type == "auth_verify") {

		return fmt.Errorf(
			"client is already authenticated",
		)
	}

	if message.Type == "auth_hello" || message.Type == "auth_verify" {
		if !client.AllowAuthAttempt(5) {
			return fmt.Errorf("authentication rate limit exceeded")
		}
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
