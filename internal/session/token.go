package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const tokenBytes = 32

func GenerateToken() (string, error) {
	data := make([]byte, tokenBytes)

	if _, err := rand.Read(data); err != nil {
		return "", fmt.Errorf(
			"failed to generate session token: %w",
			err,
		)
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}
