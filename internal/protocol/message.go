package protocol

import (
	"encoding/json"
	"fmt"
)

const Version = 1

type Message struct {
	Version int `json:"version,omitempty"`
	Type    string `json:"type"`

	ID string `json:"id,omitempty"`

	Payload json.RawMessage `json:"payload,omitempty"`

	// Fields contains all top-level fields from the incoming
	// Ether protocol message.
	Fields map[string]json.RawMessage `json:"-"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Decode(data []byte) (Message, error) {
	var raw map[string]json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return Message{}, fmt.Errorf(
			"invalid JSON: %w",
			err,
		)
	}

	var message Message

	if value, ok := raw["version"]; ok {
		if err := json.Unmarshal(
			value,
			&message.Version,
		); err != nil {
			return Message{}, fmt.Errorf(
				"invalid version",
			)
		}
	}

	if value, ok := raw["type"]; ok {
		if err := json.Unmarshal(
			value,
			&message.Type,
		); err != nil {
			return Message{}, fmt.Errorf(
				"invalid message type",
			)
		}
	}

	if message.Type == "" {
		return Message{}, fmt.Errorf(
			"missing message type",
		)
	}

	if value, ok := raw["id"]; ok {
		_ = json.Unmarshal(
			value,
			&message.ID,
		)
	}

	if value, ok := raw["payload"]; ok {
		message.Payload = value
	}

	message.Fields = raw

	return message, nil
}

func NewMessage(
	messageType string,
	payload any,
) (Message, error) {

	data, err := json.Marshal(payload)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Version: Version,
		Type:    messageType,
		Payload: data,
	}, nil
}

func (m Message) Field(
	name string,
	target any,
) error {

	data, exists := m.Fields[name]

	if !exists {
		return fmt.Errorf(
			"missing field: %s",
			name,
		)
	}

	if err := json.Unmarshal(
		data,
		target,
	); err != nil {
		return fmt.Errorf(
			"invalid field %s: %w",
			name,
			err,
		)
	}

	return nil
}

// LegacyMessage creates a message using the original
// Ether wire format:
//
// {
//   "type": "...",
//   "serverId": "..."
// }
func LegacyMessage(
	messageType string,
	fields map[string]any,
) ([]byte, error) {

	message := make(
		map[string]any,
		len(fields)+1,
	)

	message["type"] = messageType

	for key, value := range fields {
		message[key] = value
	}

	return json.Marshal(message)
}