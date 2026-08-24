package protocol

import (
	"encoding/json"
	"fmt"
)

func NewErrorMessage(code string, message string) (Message, error) {
	return NewMessage(
		"error",
		ErrorPayload{
			Code:    code,
			Message: message,
		},
	)
}

func EncodeMessage(message Message) ([]byte, error) {
	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to encode protocol message: %w", err)
	}

	return data, nil
}